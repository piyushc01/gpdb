package utils_test

import (
	"errors"
	"github.com/greenplum-db/gp-common-go-libs/testhelper"
	"net"
	"os"
	"reflect"
	"testing"

	"github.com/greenplum-db/gpdb/gp/testutils"
	"github.com/greenplum-db/gpdb/gp/utils"
)

func TestGetListDifference(t *testing.T) {
	testhelper.SetupTestLogger()
	t.Run("returns correct diff when both lists are same", func(t *testing.T) {
		list1 := []string{"s1", "s2", "s3"}
		list2 := []string{"s1", "s2", "s3"}
		var expResult []string
		result := utils.GetListDifference(list1, list2)
		if !reflect.DeepEqual(result, expResult) {
			t.Fatalf("got:%v, expected:%v", result, list1)
		}
	})
	t.Run("returns correct diff when list1 and list2 has some common elements", func(t *testing.T) {
		list1 := []string{"s1", "s2", "s3"}
		list2 := []string{"s1", "s2"}
		expResult := []string{"s3"}
		result := utils.GetListDifference(list1, list2)
		if !reflect.DeepEqual(result, expResult) {
			t.Fatalf("got:%v, expected:%v", result, list1)
		}
	})
	t.Run("returns correct diff when list2 has some extra elements", func(t *testing.T) {
		list1 := []string{"s1", "s2", "s3"}
		list2 := []string{"s2", "s1", "s4", "s5"}
		expResult := []string{"s3"}
		result := utils.GetListDifference(list1, list2)
		if !reflect.DeepEqual(result, expResult) {
			t.Fatalf("got:%v, expected:%v", result, list1)
		}
	})
	t.Run("returns correct diff when list2 has no elements", func(t *testing.T) {
		list1 := []string{"s1", "s2", "s3"}
		list2 := []string{}
		expResult := list1
		result := utils.GetListDifference(list1, list2)
		if !reflect.DeepEqual(result, expResult) {
			t.Fatalf("got:%v, expected:%v", result, list1)
		}
	})
	t.Run("returns correct diff when list2 has some extra elements", func(t *testing.T) {
		list1 := []string{}
		list2 := []string{"s1", "s2", "s4", "s5"}
		expResult := list1
		result := utils.GetListDifference(list1, list2)
		if len(expResult) != 0 {
			t.Fatalf("got:%v, expected:%v", result, list1)
		}
	})
}

func TestWriteLinesToFile(t *testing.T) {
	testhelper.SetupTestLogger()
	t.Run("succesfully writes to the file", func(t *testing.T) {
		file := "/tmp/testfile001"
		defer os.Remove(file)

		lines := []string{"line1", "line2", "line3"}
		err := utils.WriteLinesToFile(file, lines)
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}

		expected := "line1\nline2\nline3"
		testutils.AssertFileContents(t, file, expected)
	})

	t.Run("errors out when not able to create the file", func(t *testing.T) {
		file, err := os.CreateTemp("", "")
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
		defer os.Remove(file.Name())

		err = os.Chmod(file.Name(), 0000)
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}

		err = utils.WriteLinesToFile(file.Name(), []string{})

		expectedErr := os.ErrPermission
		if !errors.Is(err, expectedErr) {
			t.Fatalf("got %#v, want %#v", err, expectedErr)
		}
	})

	t.Run("errors out when not able to write to the file", func(t *testing.T) {
		utils.System.Create = func(name string) (*os.File, error) {
			_, writer, _ := os.Pipe()
			writer.Close()

			return writer, nil
		}
		defer utils.ResetSystemFunctions()

		err := utils.WriteLinesToFile("", []string{})

		expectedErr := os.ErrClosed
		if !errors.Is(err, expectedErr) {
			t.Fatalf("got %#v, want %#v", err, expectedErr)
		}
	})
}

func TestGetHostAddrsNoLoopback(t *testing.T) {
	testhelper.SetupTestLogger()
	t.Run("returns the correct address without loopback", func(t *testing.T) {
		utils.System.InterfaceAddrs = func() ([]net.Addr, error) {
			_, addr1, _ := net.ParseCIDR("192.0.1.0/24")
			_, addr2, _ := net.ParseCIDR("2001:db8::/32")
			_, loopbackAddrIp4, _ := net.ParseCIDR("127.0.0.1/8")
			_, loopbackAddrIp6, _ := net.ParseCIDR("::1/128")

			return []net.Addr{addr1, addr2, loopbackAddrIp4, loopbackAddrIp6}, nil
		}
		defer utils.ResetSystemFunctions()

		result, err := utils.GetHostAddrsNoLoopback()
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}

		expected := []string{"192.0.1.0/24", "2001:db8::/32"}
		if !reflect.DeepEqual(result, expected) {
			t.Fatalf("got %v, want %v", result, expected)
		}
	})

	t.Run("errors out when not able to get the host address", func(t *testing.T) {
		expectedErr := errors.New("error")
		utils.System.InterfaceAddrs = func() ([]net.Addr, error) {
			return nil, expectedErr
		}
		defer utils.ResetSystemFunctions()

		_, err := utils.GetHostAddrsNoLoopback()
		if !errors.Is(err, expectedErr) {
			t.Fatalf("got %#v, want %#v", err, expectedErr)
		}
	})
}
