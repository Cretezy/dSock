// Simple tests in here for simple functions
package common_test

import (
	"github.com/Cretezy/dSock/common"
	"github.com/stretchr/testify/suite"
	"testing"
)

type UtilsSuite struct {
	suite.Suite
}

func TestUtilsSuite(t *testing.T) {
	suite.Run(t, new(UtilsSuite))
}

func (suite *UtilsSuite) TestRemoveString() {
	suite.Equal(
		[]string{"b", "a"},
		common.RemoveString([]string{"a", "b", "a"}, "a"),
	)
}

func (suite *UtilsSuite) TestUniqueString() {
	suite.Equal(
		[]string{"a", "b"},
		common.UniqueString([]string{"a", "b", "a"}),
	)
}

func (suite *UtilsSuite) TestRemoveEmpty() {
	suite.Equal(
		[]string{"a", "b"},
		common.RemoveEmpty([]string{"a", "b", ""}),
	)
}

func (suite *UtilsSuite) TestRandomString() {
	suite.Len(common.RandomString(8), 8)
}
