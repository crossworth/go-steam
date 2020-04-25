package kv_test

import (
	"bytes"
	"testing"

	"github.com/13k/go-steam/kv"
	"github.com/stretchr/testify/require"
)

func TestBinaryDecoder_Decode(t *testing.T) {
	require := require.New(t)

	//nolint:lll
	testCases := []struct {
		Data     []byte
		Expected kv.KeyValue
		Err      string
	}{
		{
			Data:     nil,
			Expected: kv.NewKeyValueEmpty(),
			Err:      "EOF",
		},
		{
			Data:     []byte{},
			Expected: kv.NewKeyValueEmpty(),
			Err:      "EOF",
		},
		{
			Data:     []byte{kv.TypeEnd.Byte()},
			Expected: kv.NewKeyValue(kv.TypeEnd, "", "", nil),
		},
		{
			Data:     []byte{kv.TypeString.Byte()},
			Expected: kv.NewKeyValue(kv.TypeString, "", "", nil),
			Err:      "EOF",
		},
		{
			Data:     []byte{kv.TypeString.Byte(), 'K'},
			Expected: kv.NewKeyValue(kv.TypeString, "", "", nil),
			Err:      "EOF",
		},
		{
			Data:     []byte{kv.TypeString.Byte(), 'K', 0x00},
			Expected: kv.NewKeyValue(kv.TypeString, "K", "", nil),
			Err:      "EOF",
		},
		{
			Data:     []byte{kv.TypeString.Byte(), 'K', 0x00, 'S'},
			Expected: kv.NewKeyValue(kv.TypeString, "K", "", nil),
			Err:      "EOF",
		},
		{
			Data:     []byte{kv.TypeString.Byte(), 'K', 0x00, 'S', 0x00},
			Expected: kv.NewKeyValue(kv.TypeString, "K", "S", nil),
		},
		{
			Data:     []byte{kv.TypeInt32.Byte(), 'K', 0x00, 0x01, 0x00, 0x00},
			Expected: kv.NewKeyValue(kv.TypeInt32, "K", "", nil),
			Err:      "unexpected EOF",
		},
		{
			Data: []byte{
				kv.TypeInt32.Byte(),
				'K', 0x00,
				0x01, 0x00, 0x00, 0x00,
			},
			Expected: kv.NewKeyValue(kv.TypeInt32, "K", "1", nil),
		},
		{
			Data: []byte{
				kv.TypeObject.Byte(),
				'K', 0x00,
				0x01, 's', 0x00, 'S',
			},
			Expected: kv.NewKeyValue(kv.TypeObject, "K", "", nil),
			Err:      "EOF",
		},
		{
			Data: []byte{
				kv.TypeObject.Byte(),
				'K', 0x00,
				0x01, 's', 0x00, 'S', 0x00,
				kv.TypeEnd.Byte(),
			},
			Expected: kv.NewKeyValueRoot("K").AddString("s", "S"),
		},
		{
			Data: binData1,
			Expected: kv.NewKeyValueRoot("RP").
				AddString("status", "#DOTA_RP_LEAGUE_MATCH_PLAYING_AS").
				AddString("steam_display", "#DOTA_RP_LEAGUE_MATCH_PLAYING_AS").
				AddString("num_params", "3").
				AddString("lobby", "lobby_id: 26628083760328052 lobby_state: RUN password: true game_mode: DOTA_GAMEMODE_CM member_count: 10 max_member_count: 10 name: \"Team Secret vs Vikin.GG \" lobby_type: 1").
				AddString("party", "party_state: IN_MATCH").
				AddString("WatchableGameID", "26628083760328052").
				AddString("param0", "#DOTA_lobby_type_name_lobby").
				AddString("param1", "8").
				AddString("param2", "#npc_dota_hero_grimstroke"),
		},
		{
			Data: binData2,
			Expected: kv.NewKeyValueRoot("RP").
				AddString("status", "#DOTA_RP_PLAYING_AS").
				AddString("steam_display", "#DOTA_RP_PLAYING_AS").
				AddString("num_params", "3").
				AddString("CustomGameMode", "0").
				AddString("WatchableGameID", "26628083785444387").
				AddString("party", "party_id: 26628083781803523 party_state: IN_MATCH open: false members { steam_id: 76561198054320440 }").
				AddString("param0", "#DOTA_lobby_type_name_ranked").
				AddString("param1", "15").
				AddString("param2", "#npc_dota_hero_zuus"),
		},
		{
			Data: binData3,
			Expected: kv.NewKeyValueRoot("RP").
				AddString("status", "#DOTA_RP_PLAYING_AS").
				AddString("steam_display", "#DOTA_RP_PLAYING_AS").
				AddString("num_params", "3").
				AddString("watching_server", "[A:1:3441750017:14553]").
				AddString("watching_from_server", "[A:1:1671049217:14554]").
				AddString("party", "party_state: IN_MATCH").
				AddString("WatchableGameID", "26628083799951455").
				AddString("steam_player_group", "26628083752106249").
				AddString("steam_player_group_size", "2").
				AddString("param0", "#DOTA_lobby_type_name_ranked").
				AddString("param1", "2").
				AddString("param2", "#npc_dota_hero_rubick"),
		},
		{
			Data: binData4,
			Expected: kv.NewKeyValueRoot("RP").
				AddString("status", "#DOTA_RP_HERO_SELECTION").
				AddString("steam_display", "#DOTA_RP_HERO_SELECTION").
				AddString("num_params", "1").
				AddString("WatchableGameID", "26628083824762603").
				AddString("steam_player_group", "26628083765767134").
				AddString("steam_player_group_size", "2").
				AddString("party", "party_id: 26628083765767134 party_state: IN_MATCH open: false members { steam_id: 76561198235766844 } members { steam_id: 76561197978446698 }").
				AddString("watching_server", "[A:1:300033030:14554]").
				AddString("watching_from_server", "[A:1:1361739785:14554]").
				AddString("param0", "#DOTA_lobby_type_name_ranked"),
		},
	}

	for testCaseIdx, testCase := range testCases {
		actual := kv.NewKeyValueEmpty()
		dec := kv.NewBinaryDecoder(bytes.NewReader(testCase.Data))
		err := dec.Decode(actual)

		if testCase.Err == "" {
			require.NoErrorf(err, "test case %d", testCaseIdx)
		} else {
			require.EqualErrorf(err, testCase.Err, "test case %d", testCaseIdx)
		}

		expected := testCase.Expected

		requireEqualfKeyValue(require, expected, actual, "test case %d", testCaseIdx)
	}
}

func requireEqualfKeyValue(r *require.Assertions, expected, actual kv.KeyValue, format string, args ...interface{}) {
	r.Equalf(expected.Type(), actual.Type(), format+": wrong KeyValue type", args...)
	r.Equalf(expected.Key(), actual.Key(), format+": wrong KeyValue key", args...)
	r.Equalf(expected.Value(), actual.Value(), format+": wrong KeyValue value", args...)
	r.Lenf(actual.Children(), len(expected.Children()), format+": wrong KeyValue children count", args...)

	for childIdx, expectedChild := range expected.Children() {
		actualChild := actual.Children()[childIdx]
		requireEqualfKeyValue(r, expectedChild, actualChild, format+": child %d", append(args, childIdx)...)
	}
}
