# Proposal: Treating the Logger as a First-Class Dependency

## 1. Executive Summary

This document proposes a change to our application's logging strategy. We advocate for moving the logger from a value hidden inside `context.Context` to an explicit, first-class parameter in our function signatures.

While the community often places the logger in the context to reduce parameter boilerplate, this pattern introduces significant drawbacks in safety, clarity, and code maintenance, especially with immutable loggers like `zap`.

By making the logger an explicit parameter, we gain compile-time safety, create more honest and readable APIs, and build a more robust and maintainable system.

**Proposed Signature:**
`func DoWork(ctx context.Context, logger *zap.Logger, ...)`

**Current Signature:**
`func DoWork(ctx context.Context, ...)` (with logger retrieved from `ctx`)

---

## 2. The Challenge: Managing Request-Scoped Dependencies

In any service-oriented application, handling a request involves a chain of function calls. For effective monitoring and debugging, we need to carry request-scoped data—such as a `request_id`, `trace_id`, or `user_id`—throughout this entire chain.

The logger is the primary consumer of this data. It must be progressively enriched with new contextual fields as it's passed down through the service layer, ensuring that every log message is traceable and informative. The core architectural question is: how do we best pass and enrich this logger?

---

## 3. The Idiomatic Go Pattern: The Logger-in-Context

The most common solution in the Go community is to inject the logger into the `context.Context` at the start of a request (typically in a middleware).

```go
// Middleware injects the logger
func LoggerMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        logger := baseLogger.With(zap.String("request_id", newID()))
        ctx := context.WithValue(r.Context(), loggerKey, logger)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Downstream function retrieves it
func SomeServiceFunction(ctx context.Context, ...) {
    log := LoggerFromContext(ctx) // Helper retrieves logger from context
    log.Info("Doing work...")
}
```

**The primary motivation for this pattern is to solve the "plumbing" problem.** If a function `A` calls `B`, which calls `C`, but only `C` needs the logger, this pattern prevents `B` from needing a `logger` parameter it will never use. It keeps function signatures clean.

However, this convenience comes at a significant cost.

---

## 4. The Drawbacks of the Logger-in-Context Pattern

### 4.1. Loss of Compile-Time Safety

This is the most critical flaw. By making the logger an explicit parameter, the Go compiler guarantees that it is provided.

`func DoWork(logger *zap.Logger)`

If you forget to pass the logger, the code will not compile. This is a powerful safety net.

When the logger is hidden in the context, its existence is only checked at runtime. The standard mitigation is a helper function that returns a no-op logger if none is found. This prevents a `nil` pointer panic, but it leads to a far more insidious problem: **silent failure.** A misconfigured middleware or a developer error can cause critical production logs to be silently discarded. **A build error is always better than missing logs during a production incident.**

### 4.2. The Immutability Burden

Loggers like `zap` are immutable. Every time we add a field, a new logger instance is created.

`enrichedLogger := logger.With(zap.String("user_id", "123"))`

If we need to pass this `enrichedLogger` to a downstream function *via the context*, we must create a new context to carry it.

```go
// This verbose pattern must be repeated in every function that enriches the logger.
func ServiceA(ctx context.Context, ...) {
    log := LoggerFromContext(ctx)
    enrichedLogger := log.With(zap.String("user_id", "123"))
    newCtx := context.WithValue(ctx, loggerKey, enrichedLogger) // Create a new context
    s.serviceB.DoSomething(newCtx, ...) // Pass the new context
}
```
This is verbose and error-prone. It completely negates the "clean signature" argument and adds significant boilerplate, making a simple and common operation—enriching log context—cumbersome.

### 4.3. Dishonest Function Signatures

A function signature is a contract. `func DoWork(ctx context.Context)` implies that the function's only cross-cutting dependency is the context itself (for cancellation, deadlines, etc.).

If the function's first line is `log := LoggerFromContext(ctx)`, then the signature is dishonest. The function has a hidden, critical dependency on a logger being present in that context. This makes the code harder to understand, test, and reuse.

---

## 5. The Proposal: A First-Class Logger

We propose to treat the application logger as a first-class dependency, on par with `context.Context` and database connections.

The canonical signature for our functions that perform I/O or contain business logic should be:

`func DoWork(ctx context.Context, logger *zap.Logger, ...)`

### Rationale & Benefits

1.  **Restores Compile-Time Guarantees:** A missing logger is a build error. This is the highest level of safety we can achieve.
2.  **Embraces Immutability Naturally:** Enriching the logger becomes a simple, explicit, and elegant operation.
    ```go
    func ServiceA(ctx context.Context, logger *zap.Logger, ...) {
        enrichedLogger := logger.With(zap.String("user_id", "123"))
        enrichedLogger.Info("Work in Service A")
        s.serviceB.DoSomething(ctx, enrichedLogger, ...) // Pass the new variable directly
    }
    ```
3.  **Promotes API Honesty and Clarity:** The signature explicitly declares all critical dependencies. This makes the code self-documenting, easier to reason about, and easier for new developers to understand.
4.  **Improves Testability:** Unit tests become simpler. There is no need to construct a context with a value; we can simply pass a test logger directly into the function.
5.  **Aligns with Go's Philosophy of Explicitness:** Go eschews "magic." While the community has made a pragmatic exception for `context`, expanding its role as a generic dependency bucket moves away from the core philosophy. Explicitness leads to code that is more robust and easier to maintain in the long term.

## 6. Conclusion

While the logger-in-context pattern is popular, it trades compile-time safety for a problematic form of convenience. The issues are magnified by the use of modern immutable loggers.

By elevating the logger to an explicit, first-class parameter, we are choosing a path of robustness, clarity, and maintainability. We are using the Go compiler to its fullest potential to help us write correct software. This change will lead to a stronger, safer, and more explicit codebase.
