package main

import "testing"

func TestLoadPoetryDataset(t *testing.T) {
	poems, err := LoadPoetry("sql")
	if err != nil {
		t.Fatal(err)
	}
	authors := make(map[string]bool)
	for _, poem := range poems {
		authors[poem.Author] = true
	}
	if len(poems) != 766 || len(authors) != 12 {
		t.Fatalf("got %d poems by %d authors", len(poems), len(authors))
	}
	poems, err = LoadPoetry("sql/pyragy/pyragy.sql")
	if err != nil {
		t.Fatal(err)
	}
	if len(poems) != 130 || poems[0].Author != "Magtymguly Pyragy" || poems[0].Title != "Gürgeniň" {
		t.Fatalf("unexpected Pyragy records: count=%d", len(poems))
	}
}

func TestParsePoetrySQL(t *testing.T) {
	input := `INSERT INTO poets (name) VALUES ('Şair''iň');
INSERT INTO poems (poet_id, title, text, source) VALUES (
@poet_id, 'Başlyk', 'Ilkinji setir
Ikinji setir; INSERT INTO poems (\n\'söz\' \\ ýol', NULL);
INSERT INTO poems (poet_id, title, text, source) VALUES (
@poet_id, 'Başga', 'Ýene bir goşgy', 'Kitap');`
	poems, err := parsePoetrySQL(input)
	if err != nil {
		t.Fatal(err)
	}
	want := Poetry{Title: "Başlyk", Author: "Şair'iň", Text: "Ilkinji setir\nIkinji setir; INSERT INTO poems (\n'söz' \\ ýol"}
	if len(poems) != 2 || poems[0] != want {
		t.Fatalf("unexpected poems: %+v", poems)
	}
}

func TestPoetryErrors(t *testing.T) {
	for _, input := range []string{
		"",
		"INSERT INTO poets (name) VALUES ('Şair');",
		"INSERT INTO poets (name) VALUES ('Şair'); INSERT INTO poems (wrong) VALUES (1);",
		"INSERT INTO poets (name) VALUES ('Şair'); INSERT INTO poems (poet_id, title, text, source) VALUES (@poet_id, 'Title', 'unfinished",
	} {
		if _, err := parsePoetrySQL(input); err == nil {
			t.Fatalf("expected error for %q", input)
		}
	}
	if _, err := LoadPoetry(t.TempDir()); err == nil {
		t.Fatal("expected error for empty directory")
	}
	if _, err := LoadPoetry("missing.sql"); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestSelectPoetry(t *testing.T) {
	if got := SelectPoetry(nil); got != (Poetry{}) {
		t.Fatalf("empty selection: %+v", got)
	}
	want := Poetry{Title: "Goşgy", Author: "Şair", Text: "Setir"}
	if got := SelectPoetry([]Poetry{want}); got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}
