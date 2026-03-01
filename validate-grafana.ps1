$ScriptPath = Join-Path $PSScriptRoot "tests/scripts/powershell/validate-grafana.ps1"
& $ScriptPath @args
exit $LASTEXITCODE
