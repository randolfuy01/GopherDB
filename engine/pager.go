package engine

import (
	"encoding/binary"
	"fmt"
	"os"
	"sync"
)

const (
	PageSize    = 4096
	HeaderSize  = 64
	FooterSize  = 32
	MaxBodySize = PageSize - HeaderSize - FooterSize
)

type (
	PageID   uint64
	PageType uint8
)

const (
	PageTypeData PageType = iota
	PageTypeIndex
	PageTypeMetadata
	PageTypeOverflow
)

type PageHeader struct {
	PageID      PageID
	NextPageID  PageID
	PrevPageID  PageID
	RecordCount uint32
	FreeSpace   uint32
	Checksum    uint32
	PageType    PageType
	_           [27]byte
}

type PageFooter struct {
	Checksum      uint32
	PageIntegrity uint32
	_             [24]byte
}

type Page struct {
	Header PageHeader
	Body   []byte
	Footer PageFooter
	dirty  bool
	_      [7]byte
}

type Pager struct {
	file       *os.File
	mutex      sync.RWMutex
	pageCache  map[PageID]*Page
	maxPages   int
	nextPageID PageID
}

type PagerConfig struct {
	FilePath     string
	MaxCacheSize int
	ReadOnly     bool
}

// NewPager() creates a new pager based on specifics of the PagerConfig
func NewPager(config PagerConfig) (*Pager, error) {
	var newPagerErr error
	// Validate filepath
	if len(config.FilePath) == 0 {
		return nil, &PagerError{
			Op:  "NewPager",
			Err: fmt.Errorf("filepath cannot be empty"),
		}
	}

	var file *os.File
	if config.ReadOnly {
		file, newPagerErr = os.OpenFile(config.FilePath, os.O_RDONLY, 0)
		if newPagerErr != nil {
			return nil, &PagerError{
				Op:  "NewPager",
				Err: fmt.Errorf("unable to open file `%s`: %w", config.FilePath, newPagerErr),
			}
		}
	} else {
		file, newPagerErr = os.OpenFile(config.FilePath, os.O_RDWR|os.O_CREATE, 0644)
		if newPagerErr != nil {
			return nil, &PagerError{
				Op:  "NewPager",
				Err: fmt.Errorf("unable to open file `%s`: %w", config.FilePath, newPagerErr),
			}
		}
	}
	cache := make(map[PageID]*Page, config.MaxCacheSize)
	pager := &Pager{
		file:       file,
		pageCache:  cache,
		maxPages:   config.MaxCacheSize,
		nextPageID: 1,
	}

	return pager, nil
}

// Close closes the pager and flushes any pending writes
func (p *Pager) Close() error {
	flushErr := p.FlushAll()
	closeErr := p.file.Close()
	p.pageCache = make(map[PageID]*Page, p.maxPages)

	if flushErr != nil {
		return &PagerError{
			Op:  "ClosePager",
			Err: fmt.Errorf("unable to flush pages: %w", flushErr),
		}
	}

	if closeErr != nil {
		return &PagerError{
			Op:  "ClosePager",
			Err: fmt.Errorf("unable to close file: %w", closeErr),
		}
	}

	return nil
}

// ReadPage reads a page from disk by PageID
func (p *Pager) ReadPage(pageID PageID) (*Page, error) {
	if pageID > PageID(p.maxPages) {
		return nil, &PagerError{
			Op:  "ReadPage",
			Err: fmt.Errorf("unable to read page: %d", pageID),
		}
	}

	offset := int64(pageID) * PageSize
	file_info, errStat := p.file.Stat()
	if errStat != nil {
		return nil, &PagerError{
			Op:  "ReadPage",
			Err: fmt.Errorf("unable to get file info: %w", errStat),
		}
	}
	if offset > file_info.Size() {
		return nil, &PagerError{
			Op:  "ReadPage",
			Err: fmt.Errorf("out of bounds of file: %d", pageID),
		}
	}

	// Read the file
	buffer := make([]byte, PageSize)
	_, errRead := p.file.ReadAt(buffer, offset)
	if errRead != nil {
		return nil, &PagerError{
			Op:  "ReadPage",
			Err: fmt.Errorf("error reading file from offset: %w", errRead),
		}
	}

	// Partition the buffer
	headerComponent, errHeader := parseHeader(buffer)
	if errHeader != nil {
		return nil, &PagerError{
			Op:  "ReadPage",
			Err: fmt.Errorf("error reading header component for page %d: %w", pageID, errHeader),
		}
	}
	footerComponent, errFooter := parseFooter(buffer)
	if errFooter != nil {
		return nil, &PagerError{
			Op:  "ReadPage",
			Err: fmt.Errorf("error reading footer component for page %d: %w", pageID, errFooter),
		}
	}

	bodyComponent := buffer[HeaderSize : HeaderSize+MaxBodySize]
	if len(bodyComponent) != MaxBodySize {
		return nil, &PagerError{
			Op:  "ReadPage",
			Err: fmt.Errorf("error reading body component for page %d: %w", pageID, errFooter),
		}
	}

	page := &Page{
		Header: headerComponent,
		Body:   bodyComponent,
		Footer: footerComponent,
		dirty:  false,
	}

	return page, nil
}

func parseHeader(buffer []byte) (PageHeader, error) {
	var header PageHeader
	header.PageID = PageID(binary.LittleEndian.Uint64(buffer[0:8]))
	header.NextPageID = PageID(binary.LittleEndian.Uint64(buffer[8:16]))
	header.PrevPageID = PageID(binary.LittleEndian.Uint64(buffer[16:24]))
	header.RecordCount = binary.LittleEndian.Uint32(buffer[24:28])
	header.FreeSpace = binary.LittleEndian.Uint32(buffer[28:32])
	header.Checksum = binary.LittleEndian.Uint32(buffer[32:36])
	header.PageType = PageType(buffer[36])
	return header, nil
}

func parseFooter(buffer []byte) (PageFooter, error) {
	var footer PageFooter
	footerStart := HeaderSize + MaxBodySize
	footer.Checksum = binary.LittleEndian.Uint32(buffer[footerStart : footerStart+4])
	footer.PageIntegrity = binary.LittleEndian.Uint32(buffer[footerStart+4 : footerStart+8])
	return footer, nil
}

// WritePage writes a page to disk
func (p *Pager) WritePage(page *Page) error {
	// TODO: Implement page writing with ACID compliance
	return nil
}

// AllocatePage allocates a new page and returns its PageID
func (p *Pager) AllocatePage(pageType PageType) (*Page, error) {
	// TODO: Implement page allocation
	return nil, nil
}

// DeallocatePage marks a page as free for reuse
func (p *Pager) DeallocatePage(pageID PageID) error {
	// TODO: Implement page deallocation
	return nil
}

// FlushPage forces a page to be written to disk
func (p *Pager) FlushPage(pageID PageID) error {
	// TODO: Implement page flushing
	return nil
}

// FlushAll flushes all dirty pages to disk
func (p *Pager) FlushAll() error {
	// TODO: Implement full cache flush
	return nil
}

// GetPageCount returns the total number of pages from the pager
func (p *Pager) GetPageCount() uint64 {
	// TODO: Implement page count retrieval
	if p.pageCache != nil {
		return uint64(len(p.pageCache))
	}
	return 0
}

// ValidatePage validates the integrity of a page using checksums
func (p *Pager) ValidatePage(page *Page) error {
	// TODO: Implement page validation
	return nil
}

// NewPage creates a new page with the given type
func NewPage(pageType PageType) *Page {
	return &Page{
		Header: PageHeader{},
		Body:   make([]byte, MaxBodySize),
		Footer: PageFooter{},
		dirty:  false,
	}
}

// Error types for the pager
type PagerError struct {
	Op  string
	Err error
}

func (e *PagerError) Error() string {
	return e.Op + ": " + e.Err.Error()
}
