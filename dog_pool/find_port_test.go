package dog_pool

import "testing"
import "github.com/orfjackal/gospec/src/gospec"

func TestFindPortSpecs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in benchmark mode.")
		return
	}
	r := gospec.NewRunner()
	r.AddSpec(FindPortSpecs)
	gospec.MainGoTest(r, t)
}

func FindPortSpecs(c gospec.Context) {
	c.Specify("[findPort] Finds an open port", func() {
		port, err := findPort()
		c.Expect(err, gospec.Equals, nil)
		c.Expect(port, gospec.Satisfies, port >= 1024)
	})

}
