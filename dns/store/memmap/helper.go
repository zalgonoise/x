package memmap

func deleteAddress(m *MemoryStore, addr string) {
	for domain, rmap := range m.Records {
		for rtype, address := range rmap {
			if address == addr {
				delete(m.Records[domain], rtype)
			}
		}
	}

}
func deleteDomain(m *MemoryStore, name string) {
	for domain := range m.Records {
		if domain == name {
			delete(m.Records, domain)
		}
	}
}
func deleteDomainByType(m *MemoryStore, name string, rtype string) {
	for domain, rmap := range m.Records {
		if domain == name {
			for rt := range rmap {
				if rt == rtype {
					delete(m.Records[domain], rtype)
				}
			}
		}
	}
}
