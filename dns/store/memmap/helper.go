package memmap

func deleteAddress(m *MemoryStore, addr string) {
	for rtype, rmap := range m.Records {
		for domain, address := range rmap {
			if address == addr {
				delete(m.Records[rtype], domain)
			}
		}
	}

}
func deleteDomain(m *MemoryStore, name string) {
	for rtype, domains := range m.Records {
		for domain := range domains {
			if domain == name {
				delete(m.Records[rtype], domain)
			}
		}
	}
}
func deleteDomainByType(m *MemoryStore, name string, rtype string) {
	for recordtype, domains := range m.Records {
		if recordtype == rtype {
			for domain := range domains {
				if domain == name {
					delete(m.Records[recordtype], domain)
				}
			}
		}
	}
}
