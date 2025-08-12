package engine

import (
	"testing"
)

var TestConfigurations PagerConfig = PagerConfig{
	FilePath:     "test",
	MaxCacheSize: 100,
	ReadOnly:     false,
}

func TestPager(t *testing.T) {
	pager, err := NewPager(TestConfigurations)
	if err != nil {
		t.Errorf(`NewPager(TestConfigurations) got %q wanted nil`, err)
	}

	if pager.file.Name() != TestConfigurations.FilePath {
		t.Errorf(`pager.file.Name() = %q; want %q`, pager.file.Name(), TestConfigurations.FilePath)
	}
	if pager.maxPages != 100 {
		t.Errorf(`pager.maxPages = %d; want 100`, pager.maxPages)
	}
}

func TestCreatePage(t *testing.T) {
	page_type := PageTypeData
	page := NewPage(page_type)
	if page == nil {
		t.Errorf(`NewPage(page_type) is nil`)
	}
}
