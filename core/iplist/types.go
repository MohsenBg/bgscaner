package iplist

type IPList struct {
	IP     string
	Enable bool
}

func New(ip string, enable int) IPList {
	return IPList{
		IP:     ip,
		Enable: enable > 0,
	}
}

// EncodeCSV returns the CSV representation of the IPList.
func (r IPList) EncodeCSV() []string {
	if r.Enable {
		return []string{r.IP, "1"}
	}
	return []string{r.IP, "0"}
}
