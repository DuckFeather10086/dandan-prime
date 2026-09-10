//go:build !js && !wasm
// +build !js,!wasm

package matchusecase

import "testing"

func TestEpisodeNumber(t *testing.T) {
	cases := []struct {
		file string
		want int
		ok   bool
	}{
		{"[SumiSora&MAGI_ATELIER][Kara_no_Kyoukai][BDRip][03][x264_1080p][flac_6ch](5FFFAA6A).mkv", 3, true},
		{"[Airota&LoliHouse] GIRLS und PANZER Finale - 04 [BDRip 1080p HEVC-10bit FLAC ASSx2].mkv", 4, true},
		{"[LoliHouse] Mashle - 17 [WebRip 1080p HEVC-10bit AAC].mkv", 17, true},
		{"[Kamigami] PSYCHO-PASS S01E12 [BD 1080p x265 Ma10p AAC].mkv", 12, true},
		{"第08话 とある少女の物語.mkv", 8, true},
		{"[ANK-Raws] Hibike! Euphonium 2 - 第03話 (BDrip 1920x1080 HEVC-YUV420P10 FLAC).mkv", 3, true},
		// Must not read the resolution, bit depth, year or CRC as an episode.
		{"[VCB-Studio] Kotonoha no Niwa [Hi10p_1080p].mkv", 0, false},
		{"风筝.Kite.1998.BD.1080P.x264.AAC5.1.mkv", 0, false},
		{"[SumiSora&MAGI_ATELIER][Kara_no_Kyoukai][BDRip][Menu1][x264_2flac](5FB0EC84).mkv", 0, false},
	}
	for _, c := range cases {
		got, ok := EpisodeNumber(c.file)
		if ok != c.ok || got != c.want {
			t.Errorf("EpisodeNumber(%q) = %d,%v; want %d,%v", c.file, got, ok, c.want, c.ok)
		}
	}
}

func TestIsExtra(t *testing.T) {
	extras := []string{
		"[SumiSora&MAGI_ATELIER][Kara_no_Kyoukai][BDRip][Menu1][x264_2flac](5FB0EC84).mkv",
		"[SumiSora&MAGI_ATELIER][Kara_no_Kyoukai][BDRip][Reminder4][x264_2flac](CE9ED5B4).mkv",
		"[LoliHouse] Bocchi the Rock! - NCOP [BDRip 1080p].mkv",
		"[DBD-Raws][孤独摇滚！][特典映像][1080P].mkv",
	}
	for _, f := range extras {
		if !IsExtra(f) {
			t.Errorf("IsExtra(%q) = false; want true", f)
		}
	}
	features := []string{
		"[SumiSora&MAGI_ATELIER][Kara_no_Kyoukai][BDRip][01][x264_1080p][flac_6ch](77D836D1).mkv",
		"[VCB-Studio] Kotonoha no Niwa [Hi10p_1080p].mkv",
	}
	for _, f := range features {
		if IsExtra(f) {
			t.Errorf("IsExtra(%q) = true; want false", f)
		}
	}
}

func TestSearchKeyword(t *testing.T) {
	cases := map[string]string{
		"[Nekomoe kissaten&VCB-Studio] Heike Monogatari [Ma10p_1080p]":         "Heike Monogatari",
		"[VCB-Studio] Yuru Camp [Ma10p_1080p]":                                 "Yuru Camp",
		"[YJDL-Studio] 聲の形 (BDrip 1920x1080 HEVC-YUV420P10 FLAC DTS-HDMA SUP)": "聲の形",
		"Dungeon Meshi": "Dungeon Meshi",
		// Nothing outside the brackets: fall through to the first non-technical block.
		"[SumiSora&MAGI_ATELIER][Kara_no_Kyoukai][BDRip]": "SumiSora&MAGI ATELIER",
		"[Code Geass R1][BDRIP]":                          "Code Geass R1",
		"[2015][Ninja Slayer][BDRIP][1080P][1-26Fin+SP]":  "Ninja Slayer",
	}
	for in, want := range cases {
		if got := SearchKeyword(in); got != want {
			t.Errorf("SearchKeyword(%q) = %q; want %q", in, got, want)
		}
	}
}

func TestEpisodeNumberSurroundedByNonTechTags(t *testing.T) {
	// Release tags that are not in the technical vocabulary ("Baha", "ANi",
	// "Bilibili") survive cleaning and end up sitting right after the episode
	// number, so the number cannot require a bracket or end-of-string after it.
	cases := []struct {
		file string
		want int
	}{
		{"[ANi] BanG Dream! Ave Mujica - 01 [1080P][Baha][WEB-DL][AAC 128Kbps 2ch][CHT].mp4", 1},
		{"[ANi] 不時輕聲地以俄語遮羞的鄰座艾莉同學 - 07 [1080P][Baha][WEB-DL][AAC AVC][CHT].mp4", 7},
		{"Apocalypse Hotel - 12 [1080P][Bilibili][WEB-DL][AAC AVC][CHS&JP].mp4", 12},
	}
	for _, c := range cases {
		if got, ok := EpisodeNumber(c.file); !ok || got != c.want {
			t.Errorf("EpisodeNumber(%q) = %d,%v; want %d,true", c.file, got, ok, c.want)
		}
	}
	// Still must not read a bit-depth suffix as an episode.
	if got, ok := EpisodeNumber("[Group] Title [BDRip 1080p HEVC-10bit FLAC].mkv"); ok {
		t.Errorf("EpisodeNumber(HEVC-10bit) = %d,true; want no match", got)
	}
}

func TestEpisodeNumberWithReleaseVersion(t *testing.T) {
	// "[01v2]" is episode 1, revision 2 -- a corrected re-upload. Without
	// handling the suffix these files get no episode number at all and drop
	// out of the season listing, which is how episodes 1 and 3 of Hibike!
	// Euphonium 3 went missing while 2 and 4 showed up.
	cases := map[string]int{
		"[KitaujiSub] Hibike! Euphonium 3 [01v2][WebRip][HEVC_AAC][CHS_JP&CHT_JP].mkv": 1,
		"[KitaujiSub] Hibike! Euphonium 3 [03v2][WebRip][HEVC_AAC][CHS_JP&CHT_JP].mkv": 3,
		"[KitaujiSub] Hibike! Euphonium 3 [02][WebRip][HEVC_AAC][CHS_JP].mp4":          2,
	}
	for file, want := range cases {
		if got, ok := EpisodeNumber(file); !ok || got != want {
			t.Errorf("EpisodeNumber(%q) = %d,%v; want %d,true", file, got, ok, want)
		}
	}
}
