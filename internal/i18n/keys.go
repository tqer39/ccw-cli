// Package i18n provides language detection and translated message lookup
// for ccw's user-facing output (TUI, tips, help, CLI messages).
package i18n

// Lang identifies an active language. Only "en" and "ja" are supported.
type Lang string

// Supported language identifiers.
const (
	LangEN Lang = "en"
	LangJA Lang = "ja"
)

// Key is a stable identifier for a translatable message. Values match the
// dot-flattened path inside the locale YAML files.
type Key string

// Translation keys. Each constant maps to a leaf entry in locales/{en,ja}.yaml.
const (
	KeyTipRename      Key = "tip.rename"
	KeyTipFromPR      Key = "tip.fromPR"
	KeyTipCleanAll    Key = "tip.cleanAll"
	KeyTipPassthrough Key = "tip.passthrough"
	KeyTipResumeBadge Key = "tip.resumeBadge"
	KeyTipStatusBadge Key = "tip.statusBadge"
	KeyTipPRBadge     Key = "tip.prBadge"

	KeyHelpUsage Key = "help.usage"

	KeyPickerFooterInstallGh Key = "picker.footer.installGh"
	KeyPickerFooterTip       Key = "picker.footer.tip"

	KeyPickerActionMenu Key = "picker.action.menu"

	KeyPickerDeleteConfirm Key = "picker.delete.confirm"

	KeyPickerPruneSingle   Key = "picker.prune.single"
	KeyPickerPruneBulkHead Key = "picker.prune.bulkHead"
	KeyPickerPruneBulkLine Key = "picker.prune.bulkLine"
	KeyPickerPruneBulkFoot Key = "picker.prune.bulkFoot"

	KeyPickerBulkFilterHead Key = "picker.bulk.filterHead"
	KeyPickerBulkFilterLine Key = "picker.bulk.filterLine"
	KeyPickerBulkFilterKeys Key = "picker.bulk.filterKeys"
	KeyPickerBulkFilterFoot Key = "picker.bulk.filterFoot"

	KeyPickerBulkConfirmHead Key = "picker.bulk.confirmHead"
	KeyPickerBulkPruneNote   Key = "picker.bulk.pruneNote"
	KeyPickerBulkDirtyWarn   Key = "picker.bulk.dirtyWarn"
	KeyPickerBulkConfirmYN   Key = "picker.bulk.confirmYN"

	KeyFallbackHeader Key = "fallback.header"
	KeyFallbackLine   Key = "fallback.line"
	KeyFallbackNew    Key = "fallback.new"
	KeyFallbackQuit   Key = "fallback.quit"
	KeyFallbackPrompt Key = "fallback.prompt"

	KeySuperpowersPluginDirNotFound Key = "superpowers.warn.pluginDirNotFound"
)

// AllKeys returns every translation Key constant. Used by the parity test to
// verify that the YAML catalogs cover the full set.
func AllKeys() []Key {
	return []Key{
		KeyTipRename, KeyTipFromPR, KeyTipCleanAll, KeyTipPassthrough, KeyTipResumeBadge,
		KeyTipStatusBadge, KeyTipPRBadge,
		KeyHelpUsage,
		KeyPickerFooterInstallGh, KeyPickerFooterTip,
		KeyPickerActionMenu,
		KeyPickerDeleteConfirm,
		KeyPickerPruneSingle, KeyPickerPruneBulkHead, KeyPickerPruneBulkLine, KeyPickerPruneBulkFoot,
		KeyPickerBulkFilterHead, KeyPickerBulkFilterLine, KeyPickerBulkFilterKeys, KeyPickerBulkFilterFoot,
		KeyPickerBulkConfirmHead, KeyPickerBulkPruneNote, KeyPickerBulkDirtyWarn, KeyPickerBulkConfirmYN,
		KeyFallbackHeader, KeyFallbackLine, KeyFallbackNew, KeyFallbackQuit, KeyFallbackPrompt,
		KeySuperpowersPluginDirNotFound,
	}
}
