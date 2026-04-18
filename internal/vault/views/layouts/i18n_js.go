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

// I18nJSONBlob returns the JS-facing i18n strings as a JSON object serialized
// into a string. base.templ embeds it in a <script type="application/json">
// element; the first bootstrap script then JSON.parses it into window.WV_I18N.
func I18nJSONBlob() string {
	m := make(map[string]string, len(jsI18nKeys))
	for _, k := range jsI18nKeys {
		m[k] = i18n.T(k)
	}
	b, _ := json.Marshal(m)
	return string(b)
}
