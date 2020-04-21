package steamid_test

import (
	"testing"

	"github.com/13k/go-steam-resources/steamlang"
	"github.com/13k/go-steam/steamid"
)

func TestParseSteam2(t *testing.T) {
	testCases := []struct {
		Subject  string
		Expected steamid.SteamID
		Err      string
	}{
		{
			Subject:  "",
			Expected: 0,
			Err:      "",
		},
		{
			Subject:  "hello",
			Expected: 0,
			Err:      "",
		},
		{
			Subject:  "[U:1:69038686]",
			Expected: 0,
			Err:      "",
		},
		{
			Subject: "STEAM_1:0:34519343",
			Expected: steamid.New(
				steamlang.EAccountType_Individual,
				steamlang.EUniverse_Public,
				steamid.AccountID(69038686),
				steamid.DesktopInstance,
			),
			Err: "",
		},
		{
			Subject: "STEAM_1:1:34519343",
			Expected: steamid.New(
				steamlang.EAccountType_Individual,
				steamlang.EUniverse_Public,
				steamid.AccountID(69038687),
				steamid.DesktopInstance,
			),
			Err: "",
		},
		{
			Subject: "STEAM_2:1:34519343",
			Expected: steamid.New(
				steamlang.EAccountType_Individual,
				steamlang.EUniverse_Beta,
				steamid.AccountID(69038687),
				steamid.DesktopInstance,
			),
			Err: "",
		},
		{
			Subject:  "STEAM_1:1:12345678901234567890",
			Expected: 0,
			Err:      `strconv.ParseUint: parsing "12345678901234567890": value out of range`,
		},
		{
			Subject:  "STEAM_0:1:12345678901234567890",
			Expected: 0,
			Err:      `strconv.ParseUint: parsing "12345678901234567890": value out of range`,
		},
		{
			Subject:  "STEAM_0:0:12345678901234567890",
			Expected: 0,
			Err:      `strconv.ParseUint: parsing "12345678901234567890": value out of range`,
		},
	}

	for testCaseIdx, testCase := range testCases {
		actual, err := steamid.ParseSteam2(testCase.Subject)

		if testCase.Err == "" {
			if err != nil {
				t.Fatalf("case %d: expected nil error, got %v", testCaseIdx, err)
			}
		} else {
			if err == nil {
				t.Fatalf("case %d: expected non-nil error", testCaseIdx)
			} else if err.Error() != testCase.Err {
				t.Fatalf("case %d: expected error to be %q, got %q", testCaseIdx, testCase.Err, err.Error())
			}
		}

		if actual != testCase.Expected {
			t.Fatalf(
				"case %[1]d: expected %[2]d (%[2]s), got %[3]d (%[3]s)",
				testCaseIdx,
				testCase.Expected,
				actual,
			)
		}
	}
}

func TestParseSteam3(t *testing.T) {
	testCases := []struct {
		Subject  string
		Expected steamid.SteamID
		Err      string
	}{
		{
			Subject:  "",
			Expected: 0,
			Err:      "",
		},
		{
			Subject:  "hello",
			Expected: 0,
			Err:      "",
		},
		{
			Subject:  "STEAM_1:0:34519343",
			Expected: 0,
			Err:      "",
		},
		{
			Subject: "[U:1:69038686]",
			Expected: steamid.New(
				steamlang.EAccountType_Individual,
				steamlang.EUniverse_Public,
				steamid.AccountID(69038686),
				steamid.DesktopInstance,
			),
			Err: "",
		},
		{
			Subject: "[U:1:69038686:4]",
			Expected: steamid.New(
				steamlang.EAccountType_Individual,
				steamlang.EUniverse_Public,
				steamid.AccountID(69038686),
				steamid.WebInstance,
			),
			Err: "",
		},
		{
			Subject: "[T:1:69038686:4]",
			Expected: steamid.New(
				steamlang.EAccountType_Chat,
				steamlang.EUniverse_Public,
				steamid.AccountID(69038686),
				steamid.WebInstance,
			),
			Err: "",
		},
		{
			Subject: "[L:1:69038686:4]",
			Expected: steamid.New(
				steamlang.EAccountType_Chat,
				steamlang.EUniverse_Public,
				steamid.AccountID(69038686),
				steamid.WebInstance.SetChatFlags(steamid.ChatInstanceFlagLobby),
			),
			Err: "",
		},
		{
			Subject: "[c:1:69038686:4]",
			Expected: steamid.New(
				steamlang.EAccountType_Chat,
				steamlang.EUniverse_Public,
				steamid.AccountID(69038686),
				steamid.WebInstance.SetChatFlags(steamid.ChatInstanceFlagClan),
			),
			Err: "",
		},
		{
			Subject:  "[U:1:12345678901234567890]",
			Expected: 0,
			Err:      `strconv.ParseUint: parsing "12345678901234567890": value out of range`,
		},
		{
			Subject:  "[U:0:12345678901234567890]",
			Expected: 0,
			Err:      `strconv.ParseUint: parsing "12345678901234567890": value out of range`,
		},
		{
			Subject:  "[M:0:12345678901234567890]",
			Expected: 0,
			Err:      `strconv.ParseUint: parsing "12345678901234567890": value out of range`,
		},
	}

	for testCaseIdx, testCase := range testCases {
		actual, err := steamid.ParseSteam3(testCase.Subject)

		if testCase.Err == "" {
			if err != nil {
				t.Fatalf("case %d: expected nil error, got %v", testCaseIdx, err)
			}
		} else {
			if err == nil {
				t.Fatalf("case %d: expected non-nil error", testCaseIdx)
			} else if err.Error() != testCase.Err {
				t.Fatalf("case %d: expected error to be %q, got %q", testCaseIdx, testCase.Err, err.Error())
			}
		}

		if actual != testCase.Expected {
			t.Fatalf(
				"case %[1]d: expected %[2]d (%[2]s), got %[3]d (%[3]s)",
				testCaseIdx,
				testCase.Expected,
				actual,
			)
		}
	}
}

func TestSteamID_Steam2(t *testing.T) {
	testCases := []struct {
		Subject  steamid.SteamID
		Expected string
	}{
		{
			Subject:  0,
			Expected: "STEAM_0:0:0",
		},
		{
			Subject: steamid.New(
				steamlang.EAccountType_Individual,
				steamlang.EUniverse_Public,
				steamid.AccountID(69038686),
				steamid.DesktopInstance,
			),
			Expected: "STEAM_1:0:34519343",
		},
		{
			Subject: steamid.New(
				steamlang.EAccountType_Individual,
				steamlang.EUniverse_Public,
				steamid.AccountID(69038687),
				steamid.DesktopInstance,
			),
			Expected: "STEAM_1:1:34519343",
		},
		{
			Subject: steamid.New(
				steamlang.EAccountType_Individual,
				steamlang.EUniverse_Beta,
				steamid.AccountID(69038687),
				steamid.DesktopInstance,
			),
			Expected: "STEAM_2:1:34519343",
		},
		{
			Subject: steamid.New(
				steamlang.EAccountType_Multiseat,
				steamlang.EUniverse_Beta,
				steamid.AccountID(69038687),
				steamid.DesktopInstance,
			),
			Expected: "STEAM_2:1:34519343",
		},
		{
			Subject: steamid.New(
				steamlang.EAccountType_Multiseat,
				steamlang.EUniverse_Beta,
				steamid.AccountID(69038687),
				steamid.WebInstance,
			),
			Expected: "STEAM_2:1:34519343",
		},
	}

	for testCaseIdx, testCase := range testCases {
		actual := testCase.Subject.Steam2()

		if actual != testCase.Expected {
			t.Fatalf("case %d: expected %q, got %q", testCaseIdx, testCase.Expected, actual)
		}
	}
}

func TestSteamID_Steam3(t *testing.T) {
	testCases := []struct {
		Expected string
		Subject  steamid.SteamID
	}{
		{
			Subject:  0,
			Expected: "[I:0:0:0]",
		},
		{
			Subject: steamid.New(
				steamlang.EAccountType_Individual,
				steamlang.EUniverse_Public,
				steamid.AccountID(69038686),
				steamid.DesktopInstance,
			),
			Expected: "[U:1:69038686:1]",
		},
		{
			Subject: steamid.New(
				steamlang.EAccountType_Individual,
				steamlang.EUniverse_Public,
				steamid.AccountID(69038686),
				steamid.WebInstance,
			),
			Expected: "[U:1:69038686:4]",
		},
		{
			Subject: steamid.New(
				steamlang.EAccountType_Chat,
				steamlang.EUniverse_Public,
				steamid.AccountID(69038686),
				steamid.WebInstance,
			),
			Expected: "[T:1:69038686:4]",
		},
		{
			Subject: steamid.New(
				steamlang.EAccountType_Chat,
				steamlang.EUniverse_Beta,
				steamid.AccountID(69038686),
				steamid.WebInstance.SetChatFlags(steamid.ChatInstanceFlagLobby),
			),
			Expected: "[L:2:69038686:4]",
		},
		{
			Subject: steamid.New(
				steamlang.EAccountType_Chat,
				steamlang.EUniverse_Dev,
				steamid.AccountID(69038686),
				steamid.WebInstance.SetChatFlags(steamid.ChatInstanceFlagClan),
			),
			Expected: "[c:4:69038686:4]",
		},
	}

	for testCaseIdx, testCase := range testCases {
		actual := testCase.Subject.Steam3()

		if actual != testCase.Expected {
			t.Fatalf("case %d: expected %q, got %q", testCaseIdx, testCase.Expected, actual)
		}
	}
}
