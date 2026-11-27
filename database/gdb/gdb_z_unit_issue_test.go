// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

// Regression tests for reported ORM conversion and query-building issues.

package gdb

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/test/gtest"
)

// Test_Issue4401 checks that JSON map fields retain all registered primitive slices.
// See https://github.com/gogf/gf/issues/4401.
func Test_Issue4401(t *testing.T) {
	type Relation struct {
		Strings  []string  `json:"strings"`
		Ints     []int     `json:"ints"`
		Int32s   []int32   `json:"int32s"`
		Int64s   []int64   `json:"int64s"`
		Uints    []uint    `json:"uints"`
		Uint32s  []uint32  `json:"uint32s"`
		Uint64s  []uint64  `json:"uint64s"`
		Float32s []float32 `json:"float32s"`
		Float64s []float64 `json:"float64s"`
	}
	gtest.C(t, func(t *gtest.T) {
		var (
			expected = &Relation{
				Strings: []string{"a", "b"}, Ints: []int{1, 2},
				Int32s: []int32{1, 2}, Int64s: []int64{1, 2},
				Uints: []uint{1, 2}, Uint32s: []uint32{1, 2}, Uint64s: []uint64{1, 2},
				Float32s: []float32{1.5, 2.5}, Float64s: []float64{1.5, 2.5},
			}
			decoded map[string]any
		)
		encoded, err := json.Marshal(map[string]*Relation{"group": expected})
		t.AssertNil(err)
		t.AssertNil(json.Unmarshal(encoded, &decoded))
		for _, input := range []any{string(encoded), encoded, decoded} {
			var got struct {
				Relations map[string]*Relation `json:"relations"`
			}
			record := Record{"relations": gvar.New(input)}
			t.AssertNil(record.Struct(&got))
			t.AssertEQ(got.Relations["group"], expected)
		}
	})
}

// Test_Issue4401_SliceInputs covers raw values and SQL NULL in the slice converter.
func Test_Issue4401_SliceInputs(t *testing.T) {
	for _, input := range []any{[]any{1, 2}, "[1,2]", []byte("[1,2]"), gvar.New("[1,2]")} {
		gtest.C(t, func(t *gtest.T) {
			var got []int
			t.AssertNil(sliceTypeConverterFunc(input, reflect.ValueOf(&got).Elem()))
			t.AssertEQ(got, []int{1, 2})
		})
	}
	for _, input := range []any{nil, gvar.New(nil)} {
		gtest.C(t, func(t *gtest.T) {
			var got []int
			t.AssertNil(sliceTypeConverterFunc(input, reflect.ValueOf(&got).Elem()))
			t.AssertNil(got)
			got = []int{99}
			t.AssertNil(sliceTypeConverterFunc(input, reflect.ValueOf(&got).Elem()))
			t.AssertEQ(got, []int{99})
		})
	}
	gtest.C(t, func(t *gtest.T) {
		var got []int
		t.AssertNE(sliceTypeConverterFunc("not-json", reflect.ValueOf(&got).Elem()), nil)
		t.AssertNil(got)
	})
}

// Test_Issue4699 tests negative values for Limit/Page/Offset should be treated as zero.
// See https://github.com/gogf/gf/issues/4699
func Test_Issue4699(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// Create a base model for testing
		m := &Model{}

		// Test Limit with single negative parameter
		m1 := m.Limit(-1)
		t.AssertEQ(m1.limit, 0)

		// Test Limit with two parameters (start, limit) where both are negative
		m2 := m.Limit(-10, -5)
		t.AssertEQ(m2.start, 0)
		t.AssertEQ(m2.limit, 0)

		// Test Limit with mixed parameters (negative start, positive limit)
		m3 := m.Limit(-10, 5)
		t.AssertEQ(m3.start, 0)
		t.AssertEQ(m3.limit, 5)

		// Test Page with negative limit
		m4 := m.Page(1, -10)
		t.AssertEQ(m4.start, 0)
		t.AssertEQ(m4.limit, 0)

		// Test Page with negative limit on page 2
		m5 := m.Page(2, -10)
		t.AssertEQ(m5.start, 0) // (2-1) * 0 = 0
		t.AssertEQ(m5.limit, 0)

		// Test Offset with negative value
		m6 := m.Offset(-5)
		t.AssertEQ(m6.offset, 0)

		// Test Offset with positive value (sanity check)
		m7 := m.Offset(10)
		t.AssertEQ(m7.offset, 10)
	})
}

// Test_Issue4747 tests that Record.Struct and Result.Structs convert nested
// Result values wrapped by gvar.Var at every level.
// See https://github.com/gogf/gf/issues/4747
func Test_Issue4747(t *testing.T) {
	type BuyTrack struct {
		BuyTrackId     *uint    `json:"buy_track_id"`
		BuyTrackName   *string  `json:"buy_track_name"`
		BuyTrackWeight *float64 `json:"buy_track_weight"`
	}
	type Track struct {
		TrackId      *uint      `json:"track_id"`
		TrackWeight  *float64   `json:"track_weight"`
		BuyTrackList []BuyTrack `json:"buy_track_list"`
	}
	type Item struct {
		ItemId    *uint   `json:"item_id"`
		TrackList []Track `json:"track_list"`
	}
	var (
		buyTrackList = Result{{
			"buy_track_id":     gvar.New(1),
			"buy_track_name":   gvar.New("1111"),
			"buy_track_weight": gvar.New("2.5"),
		}}
		trackList = Result{{
			"track_id":       gvar.New(2),
			"track_weight":   gvar.New("3.5"),
			"buy_track_list": gvar.New(buyTrackList),
		}}
		item = Record{
			"item_id":    gvar.New(3),
			"track_list": gvar.New(trackList),
		}
	)
	gtest.C(t, func(t *gtest.T) {
		var info Item
		t.AssertNil(item.Struct(&info))
		t.Assert(*info.ItemId, 3)
		t.Assert(len(info.TrackList), 1)
		t.Assert(*info.TrackList[0].TrackId, 2)
		t.Assert(*info.TrackList[0].TrackWeight, 3.5)
		t.Assert(len(info.TrackList[0].BuyTrackList), 1)
		t.Assert(*info.TrackList[0].BuyTrackList[0].BuyTrackId, 1)
		t.Assert(*info.TrackList[0].BuyTrackList[0].BuyTrackName, "1111")
		t.Assert(*info.TrackList[0].BuyTrackList[0].BuyTrackWeight, 2.5)
	})
	gtest.C(t, func(t *gtest.T) {
		var list []Item
		t.AssertNil(Result{item}.Structs(&list))
		t.Assert(len(list), 1)
		t.Assert(*list[0].ItemId, 3)
		t.Assert(len(list[0].TrackList), 1)
		t.Assert(*list[0].TrackList[0].TrackWeight, 3.5)
		t.Assert(len(list[0].TrackList[0].BuyTrackList), 1)
		t.Assert(*list[0].TrackList[0].BuyTrackList[0].BuyTrackWeight, 2.5)
	})
	// Record.Map keeps the nested Result as it is.
	gtest.C(t, func(t *gtest.T) {
		_, ok := item.Map()["track_list"].(Result)
		t.Assert(ok, true)
	})
}

// Test_Issue4787 tests that Result.Structs fills map slices with the record
// values, the same as Result.List, instead of leaving empty maps.
// See https://github.com/gogf/gf/issues/4787
func Test_Issue4787(t *testing.T) {
	var result = Result{
		{"id": gvar.New(1), "name": gvar.New("john")},
		{"id": gvar.New(2), "name": gvar.New("smith")},
	}
	gtest.C(t, func(t *gtest.T) {
		var list []map[string]any
		t.AssertNil(result.Structs(&list))
		t.Assert(len(list), 2)
		t.AssertEQ(list[0]["id"], 1)
		t.AssertEQ(list[0]["name"], "john")
		t.AssertEQ(list[1]["id"], 2)
		t.AssertEQ(list[1]["name"], "smith")
	})
	gtest.C(t, func(t *gtest.T) {
		var list []*map[string]any
		t.AssertNil(result.Structs(&list))
		t.Assert(len(list), 2)
		t.AssertEQ((*list[0])["name"], "john")
		t.AssertEQ((*list[1])["name"], "smith")
	})
	gtest.C(t, func(t *gtest.T) {
		var list []map[string]string
		t.AssertNil(result.Structs(&list))
		t.Assert(len(list), 2)
		t.AssertEQ(list[0]["id"], "1")
		t.AssertEQ(list[1]["name"], "smith")
	})
}
