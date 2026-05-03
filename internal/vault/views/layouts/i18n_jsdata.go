package layouts

import (
	"encoding/json"

	"github.com/sookmook/wall-vault/internal/i18n"
)

// jsI18nKeys is the subset of i18n strings the browser-side JS in base.templ
// needs at runtime (alerts, optgroup labels, SSE status text, etc.). Keep this
// list tight — every extra key ships on every dashboard HTML response.
var jsI18nKeys = []string{
	// UI strings only referenced from JS:
	"js_user_custom_suffix",
	"js_group_free",
	"js_group_paid",
	"js_not_selected",
	"js_avatar_preview",
	"js_confirm_delete",
	"js_save_fail_fmt",
	"js_reorder_fail_fmt",
	"js_delete_fail_fmt",
	"js_sse_connected",
	"js_sse_reconnecting",
	"js_sse_off",
	// Re-exposed from the main dictionary so JS and templ share wording:
	"opt_service_default",
	"lbl_group_default",
	"lbl_group_allowed",
	"warn_stale_override_short",
}

// I18nJSONBlob returns the JS-facing i18n strings for the requested locale as a
// JSON object serialized into a string. base.templ embeds it in a
// <script type="application/json"> element; the first bootstrap script then
// JSON.parses it into window.WV_I18N.
//
// Takes lang explicitly (rather than reading the global via i18n.T) because a
// concurrent request that toggles the global locale would otherwise leak the
// wrong language into this response — the symptom that surfaced as English
// optgroup labels ("Default" / "Allowed") inside an otherwise-Korean agent
// edit slideover.
func I18nJSONBlob(lang string) string {
	m := make(map[string]string, len(jsI18nKeys))
	for _, k := range jsI18nKeys {
		m[k] = i18n.TFor(lang, k)
	}
	b, _ := json.Marshal(m)
	return string(b)
}
