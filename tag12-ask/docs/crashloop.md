# Runbook: CrashLoopBackOff
Startet ein Container immer wieder neu, die Logs des vorherigen Laufs lesen:
kubectl logs <pod> --previous. Häufige Ursachen: fehlende Env-Variable,
falsches Secret, Liveness-Probe zu streng.
