{
  runCommand,
  stashsphere-openapi,
  stashsphere-openapi-lint,
}:

runCommand "stashsphere-openapi-lint-check" { } ''
  ${stashsphere-openapi-lint}/bin/stashsphere-openapi-lint ${stashsphere-openapi}/openapi.json
  touch $out
''
