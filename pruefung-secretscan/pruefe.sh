#!/usr/bin/env bash
# pruefe.sh – bewertet die Abschlussprüfung (Lab-Secrets, nur zum Testen)
cd "$(dirname "$0")" || exit 2
pts=0
check() { # check PUNKTE NAME BEFEHL
  if eval "$3" >/dev/null 2>&1; then echo "✔ $1 P  $2"; pts=$((pts + $1))
  else echo "✘ 0 P  $2"; fi
}
fx=$(mktemp -d)
mkdir -p "$fx/app" "$fx/.git" "$fx/bin"
printf 'token: ghp_%s\n' "$(printf 'b%.0s' {1..36})" > "$fx/app/config.yaml"
printf 'A=1\nAWS_KEY=AKIA%s\n' LABEXAMPLE000000 > "$fx/app/.env"
printf 'K=AKIA%s # secretscan:ignore\n' LABEXAMPLE000000 > "$fx/ok.env"
printf 'token: ghp_%s\n' "$(printf 'c%.0s' {1..36})" > "$fx/.git/config"
printf '\0\1AKIA%s' LABEXAMPLE000000 > "$fx/bin/tool"
rm -f bin/secretscan; go build -o bin/secretscan ./cmd/secretscan 2>/dev/null
out=$(./bin/secretscan "$fx" 2>/dev/null); code=$?

check 10 "gofmt sauber" 'test -z "$(gofmt -l .)"'
check 10 "go vet + golangci-lint" "go vet ./... && ~/go-lab/bin/lint run ./..."
check 20 "Tests grün mit -race" "go test -race -count=1 ./..."
check 10 "Exit 1 bei Treffern" "test $code -eq 1"
check 10 "genau 2 Treffer" "tail -1 <<<\"\$out\" | grep -qx '2 Treffer'"
check 10 "Secrets maskiert" "test $code -eq 1 && ! grep -q LABEXAMPLE <<<\"\$out\""
check 10 ".git, Binär, ignore übersprungen" \
  "test $code -eq 1 && ! grep -qE '^(\.git|bin/|ok\.env)' <<<\"\$out\""
check 10 "JSON gültig" "./bin/secretscan -json '$fx' | jq -e 'length == 2'"
check 5 "Exit 0 ohne Treffer" "./bin/secretscan cmd"
check 5 "Exit 2 ohne Ordner" "./bin/secretscan; test \$? -eq 2"
rm -rf "$fx"
echo "Punkte: $pts von 100"
