#OmarchyInput: {
	args: string @go(Args)
	// expect_non_zero: assert the command FAILED (any non-zero exit) — the
	// CLI-rejects-X class. Mirrors plugin-command's expect_non_zero; mutually
	// exclusive with the step-level exit_status matcher (exact code).
	expect_non_zero: bool | *false @go(ExpectNonZero)
}
