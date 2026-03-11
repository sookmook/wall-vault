// Package i18n: 세계 10대 언어 지원
// 지원 언어: ko(한국어) en(English) zh(中文) es(Español) hi(हिन्दी)
//            ar(العربية) pt(Português) fr(Français) de(Deutsch) ja(日本語)
package i18n

import (
	"os"
	"strings"
)

var lang = "en"

// Supported: 지원 언어 코드 목록
var Supported = []string{"ko", "en", "zh", "es", "hi", "ar", "pt", "fr", "de", "ja"}

var messages = map[string]map[string]string{
	"ko": {
		"unknown_command":   "알 수 없는 명령",
		"starting":          "시작 중",
		"stopping":          "종료 중",
		"proxy_started":     "프록시 서버 시작됨",
		"vault_started":     "키 금고 서버 시작됨",
		"health_ok":         "정상",
		"health_err":        "오류",
		"ollama_fetching":   "Ollama 모델 목록 조회 중...",
		"ollama_found":      "Ollama 모델 발견",
		"ollama_unavailable":"Ollama 미사용 가능",
		"key_cooldown":      "키 쿨다운",
		"key_active":        "키 활성",
		"key_exhausted":     "키 소진",
		"waiting_ollama":    "Ollama 대기 중...",
		"service_not_found": "서비스를 찾을 수 없음",
		"model_changed":     "모델 변경됨",
		"config_loaded":     "설정 로드 완료",
		"config_not_found":  "설정 파일 없음 — 기본값 사용",
		"setup_welcome":     "wall-vault 설치 마법사에 오신 것을 환영합니다",
		"setup_done":        "설정 완료! wall-vault start 로 시작하세요",
	},
	"en": {
		"unknown_command":   "unknown command",
		"starting":          "starting",
		"stopping":          "stopping",
		"proxy_started":     "proxy server started",
		"vault_started":     "key vault server started",
		"health_ok":         "ok",
		"health_err":        "error",
		"ollama_fetching":   "fetching Ollama model list...",
		"ollama_found":      "Ollama models found",
		"ollama_unavailable":"Ollama not available",
		"key_cooldown":      "key on cooldown",
		"key_active":        "key active",
		"key_exhausted":     "key exhausted",
		"waiting_ollama":    "waiting for Ollama...",
		"service_not_found": "service not found",
		"model_changed":     "model changed",
		"config_loaded":     "config loaded",
		"config_not_found":  "config not found — using defaults",
		"setup_welcome":     "Welcome to wall-vault setup wizard",
		"setup_done":        "Setup complete! Run: wall-vault start",
	},
	"zh": {
		"unknown_command":   "未知命令",
		"starting":          "正在启动",
		"stopping":          "正在停止",
		"proxy_started":     "代理服务器已启动",
		"vault_started":     "密钥金库已启动",
		"health_ok":         "正常",
		"health_err":        "错误",
		"ollama_fetching":   "正在获取 Ollama 模型列表...",
		"ollama_found":      "发现 Ollama 模型",
		"ollama_unavailable":"Ollama 不可用",
		"key_cooldown":      "密钥冷却中",
		"key_active":        "密钥活跃",
		"key_exhausted":     "密钥已耗尽",
		"waiting_ollama":    "等待 Ollama...",
		"service_not_found": "找不到服务",
		"model_changed":     "模型已更改",
		"config_loaded":     "配置已加载",
		"config_not_found":  "未找到配置文件 — 使用默认值",
		"setup_welcome":     "欢迎使用 wall-vault 安装向导",
		"setup_done":        "设置完成！运行：wall-vault start",
	},
	"es": {
		"unknown_command":   "comando desconocido",
		"starting":          "iniciando",
		"stopping":          "deteniendo",
		"proxy_started":     "servidor proxy iniciado",
		"vault_started":     "almacén de claves iniciado",
		"health_ok":         "ok",
		"health_err":        "error",
		"ollama_fetching":   "obteniendo lista de modelos Ollama...",
		"ollama_found":      "modelos Ollama encontrados",
		"ollama_unavailable":"Ollama no disponible",
		"key_cooldown":      "clave en enfriamiento",
		"key_active":        "clave activa",
		"key_exhausted":     "clave agotada",
		"waiting_ollama":    "esperando Ollama...",
		"service_not_found": "servicio no encontrado",
		"model_changed":     "modelo cambiado",
		"config_loaded":     "configuración cargada",
		"config_not_found":  "configuración no encontrada — usando valores predeterminados",
		"setup_welcome":     "Bienvenido al asistente de instalación de wall-vault",
		"setup_done":        "¡Configuración completa! Ejecute: wall-vault start",
	},
	"hi": {
		"unknown_command":   "अज्ञात आदेश",
		"starting":          "शुरू हो रहा है",
		"stopping":          "रुक रहा है",
		"proxy_started":     "प्रॉक्सी सर्वर शुरू हुआ",
		"vault_started":     "की वॉल्ट शुरू हुआ",
		"health_ok":         "ठीक है",
		"health_err":        "त्रुटि",
		"ollama_fetching":   "Ollama मॉडल सूची प्राप्त हो रही है...",
		"ollama_found":      "Ollama मॉडल मिले",
		"ollama_unavailable":"Ollama उपलब्ध नहीं",
		"key_cooldown":      "की कूलडाउन पर",
		"key_active":        "की सक्रिय",
		"key_exhausted":     "की समाप्त",
		"waiting_ollama":    "Ollama की प्रतीक्षा...",
		"service_not_found": "सेवा नहीं मिली",
		"model_changed":     "मॉडल बदला गया",
		"config_loaded":     "कॉन्फ़िग लोड हुआ",
		"config_not_found":  "कॉन्फ़िग नहीं मिला — डिफ़ॉल्ट उपयोग",
		"setup_welcome":     "wall-vault सेटअप विज़ार्ड में आपका स्वागत है",
		"setup_done":        "सेटअप पूर्ण! चलाएं: wall-vault start",
	},
	"ar": {
		"unknown_command":   "أمر غير معروف",
		"starting":          "جارٍ البدء",
		"stopping":          "جارٍ الإيقاف",
		"proxy_started":     "تم بدء خادم البروكسي",
		"vault_started":     "تم بدء خزنة المفاتيح",
		"health_ok":         "جيد",
		"health_err":        "خطأ",
		"ollama_fetching":   "جارٍ جلب قائمة نماذج Ollama...",
		"ollama_found":      "تم العثور على نماذج Ollama",
		"ollama_unavailable":"Ollama غير متاح",
		"key_cooldown":      "المفتاح في فترة التبريد",
		"key_active":        "المفتاح نشط",
		"key_exhausted":     "المفتاح مستنفد",
		"waiting_ollama":    "في انتظار Ollama...",
		"service_not_found": "الخدمة غير موجودة",
		"model_changed":     "تم تغيير النموذج",
		"config_loaded":     "تم تحميل الإعدادات",
		"config_not_found":  "الإعدادات غير موجودة — استخدام الافتراضيات",
		"setup_welcome":     "مرحبًا بك في معالج إعداد wall-vault",
		"setup_done":        "اكتمل الإعداد! شغّل: wall-vault start",
	},
	"pt": {
		"unknown_command":   "comando desconhecido",
		"starting":          "iniciando",
		"stopping":          "parando",
		"proxy_started":     "servidor proxy iniciado",
		"vault_started":     "cofre de chaves iniciado",
		"health_ok":         "ok",
		"health_err":        "erro",
		"ollama_fetching":   "obtendo lista de modelos Ollama...",
		"ollama_found":      "modelos Ollama encontrados",
		"ollama_unavailable":"Ollama indisponível",
		"key_cooldown":      "chave em resfriamento",
		"key_active":        "chave ativa",
		"key_exhausted":     "chave esgotada",
		"waiting_ollama":    "aguardando Ollama...",
		"service_not_found": "serviço não encontrado",
		"model_changed":     "modelo alterado",
		"config_loaded":     "configuração carregada",
		"config_not_found":  "configuração não encontrada — usando padrões",
		"setup_welcome":     "Bem-vindo ao assistente de instalação do wall-vault",
		"setup_done":        "Configuração concluída! Execute: wall-vault start",
	},
	"fr": {
		"unknown_command":   "commande inconnue",
		"starting":          "démarrage",
		"stopping":          "arrêt",
		"proxy_started":     "serveur proxy démarré",
		"vault_started":     "coffre-fort de clés démarré",
		"health_ok":         "ok",
		"health_err":        "erreur",
		"ollama_fetching":   "récupération de la liste des modèles Ollama...",
		"ollama_found":      "modèles Ollama trouvés",
		"ollama_unavailable":"Ollama indisponible",
		"key_cooldown":      "clé en refroidissement",
		"key_active":        "clé active",
		"key_exhausted":     "clé épuisée",
		"waiting_ollama":    "en attente d'Ollama...",
		"service_not_found": "service introuvable",
		"model_changed":     "modèle modifié",
		"config_loaded":     "configuration chargée",
		"config_not_found":  "configuration introuvable — utilisation des valeurs par défaut",
		"setup_welcome":     "Bienvenue dans l'assistant d'installation wall-vault",
		"setup_done":        "Configuration terminée ! Exécutez : wall-vault start",
	},
	"de": {
		"unknown_command":   "unbekannter Befehl",
		"starting":          "wird gestartet",
		"stopping":          "wird gestoppt",
		"proxy_started":     "Proxy-Server gestartet",
		"vault_started":     "Schlüsseltresor gestartet",
		"health_ok":         "ok",
		"health_err":        "Fehler",
		"ollama_fetching":   "Ollama-Modellliste wird abgerufen...",
		"ollama_found":      "Ollama-Modelle gefunden",
		"ollama_unavailable":"Ollama nicht verfügbar",
		"key_cooldown":      "Schlüssel in Abkühlung",
		"key_active":        "Schlüssel aktiv",
		"key_exhausted":     "Schlüssel erschöpft",
		"waiting_ollama":    "Warten auf Ollama...",
		"service_not_found": "Dienst nicht gefunden",
		"model_changed":     "Modell geändert",
		"config_loaded":     "Konfiguration geladen",
		"config_not_found":  "Konfiguration nicht gefunden — Standardwerte werden verwendet",
		"setup_welcome":     "Willkommen beim wall-vault Einrichtungsassistenten",
		"setup_done":        "Einrichtung abgeschlossen! Ausführen: wall-vault start",
	},
	"ja": {
		"unknown_command":   "不明なコマンド",
		"starting":          "起動中",
		"stopping":          "停止中",
		"proxy_started":     "プロキシサーバー起動済み",
		"vault_started":     "キー金庫起動済み",
		"health_ok":         "正常",
		"health_err":        "エラー",
		"ollama_fetching":   "Ollamaモデル一覧取得中...",
		"ollama_found":      "Ollamaモデル発見",
		"ollama_unavailable":"Ollama利用不可",
		"key_cooldown":      "キークールダウン",
		"key_active":        "キーアクティブ",
		"key_exhausted":     "キー枯渇",
		"waiting_ollama":    "Ollama待機中...",
		"service_not_found": "サービスが見つかりません",
		"model_changed":     "モデル変更済み",
		"config_loaded":     "設定読み込み完了",
		"config_not_found":  "設定ファイルなし — デフォルト使用",
		"setup_welcome":     "wall-vault セットアップウィザードへようこそ",
		"setup_done":        "セットアップ完了！実行: wall-vault start",
	},
}

// Init: 환경변수 또는 시스템 로케일에서 언어 결정
func Init() {
	if v := os.Getenv("WV_LANG"); v != "" {
		SetLang(v)
		return
	}
	// Linux/macOS: LANG 환경변수
	if v := os.Getenv("LANG"); v != "" {
		SetLang(v)
		return
	}
	// Windows: USERPROFILE 경로 힌트는 없으므로 기본 en
}

// SetLang: 언어 코드 설정 (외부 패키지에서 호출 가능)
func SetLang(v string) {
	v = strings.ToLower(v)
	// 정확한 매칭 우선
	for _, code := range Supported {
		if strings.HasPrefix(v, code) {
			lang = code
			return
		}
	}
	// 로케일 별칭
	switch {
	case strings.HasPrefix(v, "zh"):
		lang = "zh"
	case strings.HasPrefix(v, "he"), strings.HasPrefix(v, "iw"):
		lang = "ar" // 히브리어 → 아랍어 폴백
	default:
		lang = "en"
	}
}

// T: 현재 언어로 메시지 반환 (영어 폴백)
func T(key string) string {
	if m, ok := messages[lang]; ok {
		if s, ok := m[key]; ok {
			return s
		}
	}
	if m, ok := messages["en"]; ok {
		if s, ok := m[key]; ok {
			return s
		}
	}
	return key
}

// Lang: 현재 언어 코드 반환
func Lang() string { return lang }
