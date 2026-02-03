package brave

import (
	"strings"
	"testing"
)

func TestPlistEscapeString(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"BraveRewardsDisabled", "BraveRewardsDisabled"},
		{"a<b", "a&lt;b"},
		{"a>b", "a&gt;b"},
		{"a&b", "a&amp;b"},
		{`a"b`, "a&quot;b"},
		{"a'b", "a&apos;b"},
		{"<key>&\"'</key>", "&lt;key&gt;&amp;&quot;&apos;&lt;/key&gt;"},
	}
	for _, tt := range tests {
		got := plistEscapeString(tt.in)
		if got != tt.want {
			t.Errorf("plistEscapeString(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestSettingsToPlistXML(t *testing.T) {
	settings := []Setting{
		{Key: "BraveRewardsDisabled", Value: true, Type: TypeBool},
		{Key: "MetricsReportingEnabled", Value: false, Type: TypeBool},
		{Key: "SafeBrowsingProtectionLevel", Value: 0, Type: TypeInteger},
		{Key: "WebRtcIPHandling", Value: "disable_non_proxied_udp", Type: TypeString},
	}
	xml := settingsToPlistXML(settings)
	if !strings.HasPrefix(xml, plistXMLHeader) {
		t.Error("XML should start with plist header")
	}
	if !strings.HasSuffix(xml, plistXMLFooter) {
		t.Error("XML should end with plist footer")
	}
	// Check escaping and structure
	if !strings.Contains(xml, "<key>BraveRewardsDisabled</key>") {
		t.Error("expected key BraveRewardsDisabled")
	}
	if !strings.Contains(xml, "<true/>") {
		t.Error("expected true for BraveRewardsDisabled")
	}
	if !strings.Contains(xml, "<false/>") {
		t.Error("expected false for MetricsReportingEnabled")
	}
	if !strings.Contains(xml, "<integer>0</integer>") {
		t.Error("expected integer 0")
	}
	if !strings.Contains(xml, "<string>disable_non_proxied_udp</string>") {
		t.Error("expected string value")
	}
	// Escaping: value with & should be escaped
	settings2 := []Setting{{Key: "K", Value: "a&b", Type: TypeString}}
	xml2 := settingsToPlistXML(settings2)
	if strings.Contains(xml2, "a&b") && !strings.Contains(xml2, "a&amp;b") {
		t.Error("expected & to be escaped in string value")
	}
}

func TestIsMacOS(t *testing.T) {
	// Just ensure it doesn't panic; actual value depends on runtime
	_ = IsMacOS()
}

func TestBraveInstalled(t *testing.T) {
	// Just ensure it doesn't panic
	_ = BraveInstalled()
}
