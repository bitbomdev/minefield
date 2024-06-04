package pkg

type unionFind struct {
	count   int
	parents []uint32
}

func (this *unionFind) Find(x uint32) uint32 {
	if x == this.parents[x] {
		return x
	}
	this.parents[x] = this.Find(this.parents[x])
	return this.parents[x]
}

func (this *unionFind) Union(x, y uint32) {
	rx, ry := this.Find(x), this.Find(y)
	if rx != ry {
		this.parents[rx] = ry
		this.count--
	}
}
