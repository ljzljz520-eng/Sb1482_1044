# BUG_REPRO

The following failures were observed while validating the initial project state.
Each section records what failed, how to reproduce it, and the complete command output.
They are preserved intentionally; only failing build gates are omitted from the generated Dockerfile.

## Failure 1: Go test (.)

- Observed problem: `Go test (.)` failed in the initial project state.
- Working directory: `.`
- Command: `cd /app && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -count=1 ./...`
- Exit status: `1`

```text
?   	example.com/materialconsole/cmd/material-console	[no test files]
?   	example.com/materialconsole/internal/httpapi	[no test files]
ok  	example.com/materialconsole/internal/catalog	0.041s
ok  	example.com/materialconsole/internal/model	0.005s
ok  	example.com/materialconsole/internal/report	0.006s
ok  	example.com/materialconsole/internal/review	0.040s
ok  	example.com/materialconsole/internal/script	0.043s
--- FAIL: TestBusiness009Regression (0.03s)
    concurrent_test.go:33: concurrent update lost data: {ID:I-109 ProjectID:bug-009 Kind:video Title:更新后的标题 Source:ref://one Channel:video Status:ready Note:初始备注 Sequence:1 Approved:false}
FAIL
FAIL	example.com/materialconsole/internal/service	0.087s
ok  	example.com/materialconsole/internal/store	0.050s
ok  	example.com/materialconsole/internal/timeline	0.040s
FAIL
```

## Architecture reproduction

### linux/amd64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/material-console): exit `0`
### linux/arm64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/material-console): exit `0`
