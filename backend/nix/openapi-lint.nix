{
  writeShellApplication,
  vacuum-go,
}:

writeShellApplication {
  name = "stashsphere-openapi-lint";

  runtimeInputs = [ vacuum-go ];

  text = ''
    if [ "$#" -lt 1 ]; then
      echo "usage: stashsphere-openapi-lint <openapi.json> [extra vacuum lint flags]" >&2
      exit 2
    fi
    spec="$1"
    shift

    style=()
    if [ ! -t 1 ]; then
      style=(--no-style)
    fi

    # vacuum wants a writable HOME (not available in the nix sandbox)
    if [ ! -w "''${HOME:-/homeless-shelter}" ]; then
      HOME="$(mktemp -d)"
      export HOME
    fi

    exec vacuum lint \
      --no-banner \
      --no-update-check \
      --details \
      --all-results \
      --no-clip \
      --min-score 0 \
      --fail-severity error \
      --ruleset ${./openapi-ruleset.yaml} \
      "''${style[@]}" \
      "$@" \
      "$spec"
  '';
}
