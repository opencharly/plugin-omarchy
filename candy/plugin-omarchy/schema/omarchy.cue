#OmarchyInput: {
	// method — the omarchy surface to dispatch. "cli" (default) runs
	// `omarchy <args>` (the command center); the shell-ipc methods run the
	// omarchy-shell IPC (the SAME surface the menu and hotkeys use); the
	// cli-read methods run the omarchy-* bin commands (the acceptance
	// surfaces the beds assert).
	// S3: the words are rows of the `_method` table at the foot of the schema, and
	// the enum this field accepts is derived from that table's word list, so the
	// catalog has ONE home — a new method is one row there, and the table, the
	// derived list and the enum cannot disagree. The `*"cli"` default stays written
	// inline: S3 permits a literal default, and the default is a VALUE choice, not a
	// word the table decides.
	method: *"cli" | or(_methodWord)
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

// --- vocabulary tables (S3) --------------------------------------------------
// Every closed set of words is a table keyed by word, and its word list is
// derived from the table and used as `or()` of that list (§10.1 S3, D-PAT-3).
// The table and the enum it feeds therefore cannot disagree, and a new word is
// ONE row here. `@vocab` names the owner (§3.1): this set is compiled into THIS
// plugin, so the owner is the plugin's own word `omarchy`.
_method: {
	// the command center — the default method
	cli: {}
	// the omarchy-shell IPC — the SAME surface the menu and hotkeys use
	"shell-ping": {}
	"shell-summon": {}
	"shell-hide": {}
	"shell-list-plugins": {}
	"shell-reload-config": {}
	"shell-notifications-dismiss": {}
	"shell-notifications-send": {}
	// the omarchy-* bin commands — the acceptance surfaces the beds assert
	"channel-current": {}
	"default-browser": {}
	"default-terminal": {}
	"default-editor": {}
	"theme-current": {}
	"theme-bg-current": {}
	"font-current": {}
	"weather-location": {}
	version: {}
} @vocab(omarchy)
_methodWord: [for w, _ in _method {w}]
