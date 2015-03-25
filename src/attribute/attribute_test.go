package attribute

import (
	"testing"
)

func TestTimeRange1(t *testing.T) {
	var tr TimeRange
	tr.Init()
	if ok, err := tr.CoverToday(); !ok {
		t.Error(err)
	}
}

func TestTimeRange2(t *testing.T) {
	var tr TimeRange
	tr.Init()
	tr.AddStart(7)
	if ok, err := tr.CoverTime(6); ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(7); !ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(8); !ok {
		t.Error(err)
	}
}

func TestTimeRange3(t *testing.T) {
	var tr TimeRange
	tr.Init()
	tr.AddEnd(7)
	if ok, err := tr.CoverTime(6); !ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(7); !ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(8); ok {
		t.Error(err)
	}
}

func TestTimeRange4(t *testing.T) {
	var tr TimeRange
	tr.Init()
	tr.AddStart(7)
	tr.AddEnd(7)
	if ok, err := tr.CoverTime(6); ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(8); ok {
		t.Error(err)
	}
}

func TestTimeRange5(t *testing.T) {
	var tr TimeRange
	tr.Init()
	tr.AddStart(7)
	tr.AddEnd(88)
	if ok, err := tr.CoverTime(6); ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(18); !ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(88); !ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(89); ok {
		t.Error(err)
	}
}

func TestTimeRange6(t *testing.T) {
	var tr TimeRange
	tr.Init()
	tr.AddStart(88)
	tr.AddEnd(8)
	if ok, err := tr.CoverTime(6); !ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(18); ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(87); ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(88); !ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(89); !ok {
		t.Error(err)
	}
}

func TestTimeRange7(t *testing.T) {
	var tr TimeRange
	tr.Init()
	tr.AddStart(8)
	tr.AddEnd(18)
	tr.AddStart(28)
	tr.AddEnd(38)
	tr.AddStart(68)
	tr.AddEnd(78)
	if ok, err := tr.CoverTime(6); ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(8); !ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(9); !ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(18); !ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(19); ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(28); !ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(34); !ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(38); !ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(40); ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(68); !ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(68); !ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(72); !ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(78); !ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(79); ok {
		t.Error(err)
	}
}

func TestTimeRange8(t *testing.T) {
	var tr TimeRange
	tr.Init()
	tr.AddEnd(18)
	tr.AddStart(28)
	tr.AddEnd(38)
	tr.AddStart(68)
	tr.AddEnd(78)
	tr.AddStart(98)

	if ok, err := tr.CoverTime(9); !ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(18); !ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(19); ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(28); !ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(34); !ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(38); !ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(40); ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(68); !ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(68); !ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(72); !ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(78); !ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(79); ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(97); ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(98); !ok {
		t.Error(err)
	}
	if ok, err := tr.CoverTime(198); !ok {
		t.Error(err)
	}
}
