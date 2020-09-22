package service_test

import (
	"github.com/ai-battleground/go-basics/pkg/service"
	"github.com/smartystreets/assertions/should"
	"github.com/smartystreets/gunit"
	"testing"
)

func TestServiceFixture(t *testing.T) {
	gunit.Run(new(ServiceFixture), t)
}

type ServiceFixture struct {
	*gunit.Fixture
}

func (this *ServiceFixture) TestNewReturnsValueReady() {
	s := service.New("idle")

	this.So(s.Log, should.NotBeNil)
	this.So(s.Metadata, should.NotBeNil)
	this.So(s.Name, should.Equal, "idle")
}

func (this *ServiceFixture) TestChildReturnsValueReady() {
	parent := service.New("data")
	child := parent.Child("redis")
	this.So(child.Log, should.NotBeNil)
	this.So(child.Metadata, should.NotBeNil)
	this.So(child.Name, should.Equal, "redis")
}

func (this *ServiceFixture) TestTestServiceName() {
	this.So(service.TestService().Name, should.Equal, "test")
}
