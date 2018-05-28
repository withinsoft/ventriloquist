package main

import "testing"

func TestCheckAvatarURL(t *testing.T) {
	cases := []struct {
		url string
		err bool
	}{
		{
			url: "https://media.discordapp.net/attachments/343646476360089600/443116023030218752/IMG_20180507_032654.jpg?width=459&height=459",
		},
		{
			url: "food",
			err: true,
		},
		{
			url: "http://example.com",
			err: true,
		},
	}

	for _, cs := range cases {
		t.Run(cs.url, func(t *testing.T) {
			err := checkAvatarURL(cs.url)

			if err == nil && cs.err {
				t.Fatal("expected error but got none")
			}

			if err != nil && !cs.err {
				t.Fatalf("expected no error but got: %v", err)
			}
		})
	}
}
