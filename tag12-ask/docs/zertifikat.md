# Runbook: Zertifikat abgelaufen
Bei TLS-Fehlern "certificate has expired" das Ablaufdatum prüfen:
openssl x509 -enddate -noout -in tls.crt. cert-manager erneuert per
kubectl annotate certificate <name> cert-manager.io/issue-temporary-certificate.
