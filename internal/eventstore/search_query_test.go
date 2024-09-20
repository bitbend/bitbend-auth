package eventstore

import (
	"testing"
)

func TestNewSearchQueryBuilder(t *testing.T) {
	tenantId := TenantId("tenant_01J81HPMKFAN9TJGGWGQ11SYWQ")
	//tenantIds := []string{
	//	"tenant_01J81HR7KCHTVFD885E2SWH10T",
	//	"tenant_01J81HS8432KC7VAJTV9VW01JR",
	//	"tenant_01J81HSV236W9Q7DKXQQKMS47Y",
	//	"tenant_01J81HTE6YPHE3EGGFQ395GFY0",
	//	"tenant_01J81HTPD7KE77NP1PJN9K7411",
	//}
	resourceOwner := "organization_01J81HWGA2134EXM5F2S7XTP8Q"
	//creator := "user_01J81HYD8S7M230T5DCKK417AE"
	//createdAfter, err := time.Parse("2006-01-02 15:04:05", "2024-09-10 08:47:30")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//createdBefore, err := time.Parse("2006-01-02 15:04:05", "2024-09-15 08:47:30")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//positionGreaterEqual := decimal.NewFromFloat(10.12437896)
	//sequenceGreaterEqual := uint64(10)
	//limit := uint64(100)
	//offset := uint64(0)

	baseQuery := NewSearchQueryBuilder().
		TenantIds(tenantId).
		ResourceOwners(resourceOwner).
		AddQuery().
		AggregateTypes("user").
		AggregateIds("user_01J826K9JQJ7YCPGDY8J9J8DCQ").
		EventTypes("user.created").
		Build()

	stmt, args, err := baseQuery.ToSql()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(stmt, args)
}
