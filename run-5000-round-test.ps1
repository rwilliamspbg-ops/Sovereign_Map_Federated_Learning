$ScriptPath = Join-Path $PSScriptRoot "tests/scripts/powershell/run-5000-round-test.ps1"
& $ScriptPath @args
exit $LASTEXITCODE
