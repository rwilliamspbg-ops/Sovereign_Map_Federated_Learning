$ScriptPath = Join-Path $PSScriptRoot "tests/scripts/powershell/continuous-load-test.ps1"
& $ScriptPath @args
exit $LASTEXITCODE
