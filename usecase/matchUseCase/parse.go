//go:build !js && !wasm
// +build !js,!wasm

package matchusecase

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Filename parsing here is deliberately shallow.
//
// The library is full of release-group naming schemes and every rule-based
// title extractor tried against it leaked somewhere: picking the "longest
// bracket block" yields the group name ("Nekomoe kissaten&VCB-Studio"), and
// filtering technical tags yields the leftovers ("HEVC-YUV420P10"). So the
// keyword produced here is only used to *fetch candidates* from bangumi.tv —
// being wrong costs one extra round trip, never a wrong match, because the
// LLM is shown the raw directory name and decides from that. Correctness does
// not live in this file.
//
// Episode *numbers* are different: they are unambiguous once the technical
// tokens are out of the way, so they stay rule-based and cost nothing.

var (
	bracketRe = regexp.MustCompile(`[\[\(【（]([^\[\]\(\)【】（）]*)[\]\)】）]`)

	// Tokens that describe the encode rather than the work.
	techRe = regexp.MustCompile(`(?i)^(bdrips?|bd|bdbox|bd-box|bluray|blu-ray|webrips?|web-?dl|webdl|dvdrips?|dvd|remux|` +
		`hevc|avc|x26[45]|h\.?26[45]|hi10p?|ma10p|yuv420p10|\d{1,2}-?bits?|10bit|8bit|` +
		`flac|flacx?\d*|aac|aacx?\d*|ac3|eac3|dts|dts-hd|dts-hdma\w*|truehd|opus|mp3|\d+ch|` +
		`mkv|mp4|m4a|mka|ass|assx?\d*|pgs|sup|srt|sc|tc|` +
		`\d{3,4}[pi]|\d{3,4}x\d{3,4}|4k|2k|` +
		`simplified|traditional|chs|cht|chs&jp|cht&jp|gb|big5|jp|jpn|jpsc|jptc|eng?|` +
		`简|繁|日|英|简体|繁體|繁体|日语|英语|双语|中日|简繁|简日|繁日|内封|外挂|内嵌|字幕|无字|` +
		`fin|end|complete|repack|v\d|menu\d*|reminder\d*|ncop\d*|nced\d*|op\d*|ed\d*|cm\d*|pv\d*|` +
		`fonts?|scans?|alternative_?audio)$`)

	// Non-feature content: menus, creditless openings, previews, extras.
	//
	// Every keyword allows a trailing index ("Preview1", "Trailer5",
	// "Voice Message03"): release groups number their extras, and \b does not
	// match between a letter and a digit, so a bare \bpreview\b misses all of
	// them.
	extraRe = regexp.MustCompile(`(?i)(\bmenu\d*\b|\breminder\d*\b|\bncop\d*\b|\bnced\d*\b|` +
		`\bpreview\d*\b|\bcm\d*\b|\bpv\d*\b|\btrailer\d*\b|\bteaser\d*\b|\bmessage\d*\b|` +
		`\bcreditless\b|\binterview\d*\b|\bmaking\b|\blogo\b|\baudio guide\b|\bstage greeting\b|` +
		`特典|映像特典|菜单|预告|花絮|无字幕op|无字幕ed|drama|sound ?track|\bost\b)`)

	epPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\bS\d{1,2}E(\d{1,4})\b`),
		regexp.MustCompile(`(?i)\bEP?[\s._-]?(\d{1,4})\b`),
		regexp.MustCompile(`第\s*(\d{1,4})\s*[话話集期]`),
		// " - 01 " with whitespace on both sides of the number. The spaces
		// are load-bearing: they keep this from reading the "10" out of a
		// leftover "HEVC-10bit", which has no space after the dash.
		regexp.MustCompile(`[-–—]\s+(\d{1,4})(?:\s|$|\.)`),
		regexp.MustCompile(`^\s*(\d{1,4})\s*[-–—.\s]`),
	}

	yearRe = regexp.MustCompile(`^(19|20)\d{2}$`)
	crcRe  = regexp.MustCompile(`(?i)^[0-9a-f]{8}$`)
	numRe  = regexp.MustCompile(`^\d{1,4}$`)
)

func isTech(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return true
	}
	for _, tok := range strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == '_' || r == '+' || r == '&'
	}) {
		if !isTechToken(strings.TrimSpace(tok)) {
			return false
		}
	}
	return true
}

// isTechToken accepts a token that is technical as a whole ("WEB-DL",
// "DTS-HDMA") and also one whose hyphen-separated parts all are
// ("HEVC-10bit"), which release groups combine freely.
func isTechToken(tok string) bool {
	if tok == "" {
		return true
	}
	if techRe.MatchString(tok) {
		return true
	}
	if !strings.Contains(tok, "-") {
		return false
	}
	for _, part := range strings.Split(tok, "-") {
		if part == "" || !techRe.MatchString(part) {
			return false
		}
	}
	return true
}

// SearchKeywords turns a directory name into every string worth sending to
// bangumi.tv search, best guess first.
//
// It returns several on purpose. Deciding which bracket block holds the title
// is not reliably decidable: "[XK SPIRITS][KAMEN RIDER ZEZTZ]" puts the
// release group first and the title second, while "[Code Geass R1][BDRIP]"
// puts the title first. Rather than guess, the caller searches each keyword
// and pools the candidates, leaving the choice to the model -- which is where
// the decision belongs anyway, since it also gets shown the raw directory name.
func SearchKeywords(dirName string) []string {
	name := filepath.Base(dirName)

	var keywords []string
	add := func(candidate string) {
		c := strings.Trim(collapse(candidate), " -_·、,.")
		if c == "" || isTech(c) || yearRe.MatchString(c) || crcRe.MatchString(c) || numRe.MatchString(c) {
			return
		}
		for _, existing := range keywords {
			if existing == c {
				return
			}
		}
		keywords = append(keywords, c)
	}

	// Text outside the brackets is the strongest signal when it exists.
	add(bracketRe.ReplaceAllString(name, " "))

	for _, block := range bracketRe.FindAllStringSubmatch(name, -1) {
		add(block[1])
	}

	if len(keywords) == 0 {
		keywords = append(keywords, collapse(name))
	}
	return keywords
}

// SearchKeyword returns the best single guess. Prefer SearchKeywords.
func SearchKeyword(dirName string) string {
	return SearchKeywords(dirName)[0]
}

// IsExtra reports whether a file is bonus material rather than a feature. It
// keeps menus and creditless openings out of a collection mapping, where each
// remaining file is supposed to be one film.
func IsExtra(fileName string) bool {
	return extraRe.MatchString(filepath.Base(fileName))
}

// EpisodeNumber pulls an episode number out of a filename, ignoring the
// resolution, bit depth, year and CRC32 numbers that surround it.
func EpisodeNumber(fileName string) (int, bool) {
	base := strings.TrimSuffix(filepath.Base(fileName), filepath.Ext(fileName))

	// Drop bracket blocks that are technical, keeping bare-number blocks such
	// as "[01]" -- those are the ones worth reading.
	cleaned := bracketRe.ReplaceAllStringFunc(base, func(m string) string {
		inner := strings.Trim(m, "[]()【】（）")
		t := strings.TrimSpace(inner)
		if numRe.MatchString(t) && !yearRe.MatchString(t) {
			return " - " + t + " "
		}
		if isTech(inner) || crcRe.MatchString(t) || yearRe.MatchString(t) {
			return " "
		}
		return " " + inner + " "
	})
	cleaned = collapse(cleaned)

	for _, re := range epPatterns {
		if m := re.FindStringSubmatch(cleaned); m != nil {
			if n, err := strconv.Atoi(m[1]); err == nil && n > 0 && n < 2000 {
				return n, true
			}
		}
	}
	return 0, false
}

func collapse(s string) string {
	return strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(strings.ReplaceAll(s, "_", " "), " "))
}
