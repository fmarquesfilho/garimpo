namespace Garimpei.Infrastructure.Sources;

/// <summary>
/// Simple circuit breaker for the Cache Sidecar (L2).
/// States: Closed (normal) → Open (3 failures) → HalfOpen (after 10s) → Closed.
/// No external dependencies — manual implementation per ADR decision.
/// </summary>
public sealed class CacheCircuitBreaker
{
    private enum State { Closed, Open, HalfOpen }

    private State _state = State.Closed;
    private int _failureCount;
    private DateTime _lastFailureAt = DateTime.MinValue;
    private readonly object _lock = new();

    private const int FailureThreshold = 3;
    private static readonly TimeSpan RecoveryTimeout = TimeSpan.FromSeconds(10);

    /// <summary>
    /// True when the circuit is open and calls should bypass the cache.
    /// </summary>
    public bool IsOpen
    {
        get
        {
            lock (_lock)
            {
                if (_state == State.Open)
                {
                    if (DateTime.UtcNow - _lastFailureAt >= RecoveryTimeout)
                    {
                        _state = State.HalfOpen;
                        return false; // Allow one probe request
                    }
                    return true;
                }
                return false;
            }
        }
    }

    /// <summary>
    /// Report a successful call — resets the circuit to Closed.
    /// </summary>
    public void RecordSuccess()
    {
        lock (_lock)
        {
            _state = State.Closed;
            _failureCount = 0;
        }
    }

    /// <summary>
    /// Report a failed call — may trip the circuit to Open.
    /// </summary>
    public void RecordFailure()
    {
        lock (_lock)
        {
            _failureCount++;
            _lastFailureAt = DateTime.UtcNow;

            if (_failureCount >= FailureThreshold)
            {
                _state = State.Open;
            }
        }
    }

    /// <summary>
    /// Current state as a string for observability headers.
    /// </summary>
    public string CurrentState
    {
        get
        {
            lock (_lock)
            {
                return _state.ToString().ToLowerInvariant();
            }
        }
    }
}
