package testhelper

import (
    "fmt"
    "syscall"
    "testing"
    "time"
)

const maxConcurrency = 10
const mutexTimeout = 30 * time.Second

// Lock acquires an inter-process lock for tests that depend on external
// services (e.g., MySQL/Elasticsearch) to avoid parallel conflicts.
// It tries up to maxConcurrency distinct lock files and registers an
// unlock via t.Cleanup. If it cannot acquire within mutexTimeout, it fails.
func Lock(t *testing.T) {
    t.Helper()
    deadline := time.Now().Add(mutexTimeout)
    var fd int
    var idx int
    for {
        if time.Now().After(deadline) {
            t.Fatalf("test mutex: timeout acquiring lock within %s", mutexTimeout)
        }
        for i := 1; i <= maxConcurrency; i++ {
            f, err := tryLock(i)
            if err == nil {
                fd = f
                idx = i
                goto locked
            }
            // if EWOULDBLOCK, just continue to next
        }
        time.Sleep(50 * time.Millisecond)
    }
locked:
    t.Logf("test mutex: acquired slot %d", idx)
    t.Cleanup(func() {
        // unlock and close
        _ = syscall.Flock(fd, syscall.LOCK_UN)
        _ = syscall.Close(fd)
    })
}

func tryLock(i int) (int, error) {
    path := lockfile(i)
    fd, err := syscall.Open(path, syscall.O_CREAT|syscall.O_RDWR, 0o750)
    if err != nil {
        return -1, err
    }
    if err := syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
        _ = syscall.Close(fd)
        return -1, err
    }
    // Write pid (best-effort) so it’s easier to debug locks left behind.
    // Ignore errors; not essential for the lock to work.
    _, _ = syscall.Write(fd, []byte(fmt.Sprintf("pid=%d\n", syscall.Getpid())))
    // Seek back to start for any tooling; ignore errors.
    _, _ = syscall.Seek(fd, 0, 0)
    return fd, nil
}

func lockfile(i int) string {
    return fmt.Sprintf("/tmp/test_mutex_%d.lock", i)
}

// Unlock is not exposed; cleanup is registered automatically in Lock.
// If you need manual control, consider extending this helper to return a guard.
