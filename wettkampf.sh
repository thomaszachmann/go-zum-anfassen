# Helfer für Tag 15 – laden mit: source ~/go-lab/wettkampf.sh
start() { # start TOOL PAKET: Modul anlegen und Stoppuhr starten
  mkdir -p ~/go-lab/tag15-"$1"/{cmd/"$1",internal/"$2"} &&
    cd ~/go-lab/tag15-"$1" || return
  [ -f go.mod ] || go mod init example.com/"$1"
  printf 'version: "2"\nlinters:\n  exclusions:\n    presets: [%s]\n' \
    std-error-handling > .golangci.yml
  date +%s > .start
}
ziel() { # ziel: alle Prüfungen; nur wenn alles grün ist, zählt die Zeit
  test -z "$(gofmt -l .)" || { gofmt -l .; return 1; }
  go vet ./... && go test -race -count=1 ./... && ~/go-lab/bin/lint run ./... ||
    return 1
  echo "Zeit: $(( ($(date +%s) - $(cat .start)) / 60 )) min"
}
