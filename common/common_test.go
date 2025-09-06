package common

import (
	"testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCommon(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Common Suite")
}

var _ = Describe("Constants", func() {
	It("should have correct JSON application type", func() {
		Expect(JSONApplicationType).To(Equal("application/json"))
	})
	
	It("should have correct status constants", func() {
		Expect(FAILED).To(Equal("FAILED"))
		Expect(SUCCESS).To(Equal("SUCCESS"))
	})
	
	It("should have correct database type", func() {
		Expect(MYSQL).To(Equal("MySQL"))
	})
	
	It("should have correct admin panel URL constant", func() {
		Expect(ADMIN_PANEL_URL).To(Equal("ADMIN_PANEL_URL"))
	})
})

var _ = Describe("Permission Enum", func() {
	It("should have correct permission values", func() {
		Expect(int(OWNER)).To(Equal(0))
		Expect(int(ADMIN)).To(Equal(1))
	})
	
	It("should have correct permission ordering", func() {
		Expect(int(OWNER)).To(BeNumerically("<", int(ADMIN)))
	})
})

var _ = Describe("Status Enum", func() {
	It("should have correct status values", func() {
		Expect(int(PENDING)).To(Equal(0))
		Expect(int(APPROVED)).To(Equal(1))
		Expect(int(REJECTED)).To(Equal(2))
	})
	
	It("should have correct status ordering", func() {
		Expect(int(PENDING)).To(BeNumerically("<", int(APPROVED)))
		Expect(int(APPROVED)).To(BeNumerically("<", int(REJECTED)))
	})
})

var _ = Describe("Enum Usage", func() {
	It("should allow permission assignment", func() {
		userPermission := OWNER
		Expect(userPermission).To(Equal(OWNER))
		
		adminPermission := ADMIN
		Expect(adminPermission).To(Equal(ADMIN))
	})
	
	It("should allow status assignment", func() {
		requestStatus := PENDING
		Expect(requestStatus).To(Equal(PENDING))
		
		approvedStatus := APPROVED
		Expect(approvedStatus).To(Equal(APPROVED))
	})
})
