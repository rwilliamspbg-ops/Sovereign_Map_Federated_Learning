$ScriptPath = Join-Path $PSScriptRoot "tests/scripts/powershell/run-test.ps1"
& $ScriptPath @args
exit $LASTEXITCODE
