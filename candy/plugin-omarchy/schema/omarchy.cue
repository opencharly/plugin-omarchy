#OmarchyInput: {
	// method — the omarchy surface to dispatch. "cli" (default) runs
	// `omarchy <args>` (the command center); the shell-ipc methods run the
	// omarchy-shell IPC (the SAME surface the menu and hotkeys use); the
	// cli-read methods run the omarchy-* bin commands (the acceptance
	// surfaces the beds assert).
	method: "cli" | "shell-ping" | "shell-summon" | "shell-hide" |
		"shell-list-plugins" | "shell-reload-config" |
		"shell-notifications-dismiss" | "shell-notifications-send" |
		"channel-current" | "default-browser" | "default-terminal" |
		"default-editor" | "theme-current" | "theme-bg-current" |
		"font-current" | "weather-location" | "version" | *"cli"
	// args — the omarchy CLI args (method: cli) or the weather-location
	// args (method: weather-location). Required for those two methods.
	args?: string @go(Args)
	// plugin — the omarchy.<plugin> id for shell-summon / shell-hide.
	plugin?: string @go(Plugin)
	// payload — the JSON payload for shell-summon (optional).
	payload?: string @go(Payload)
	// title / text — the notification title/body for shell-notifications-send.
	title?: string @go(Title)
	text?: string @go(Text)
	// expect_non_zero: assert the command FAILED (any non-zero exit) — the
	// CLI-rejects-X class. Mirrors plugin-command's expect_non_zero; mutually
	// exclusive with the step-level exit_status matcher (exact code).
	expect_non_zero: bool | *false @go(ExpectNonZero)
}
