// Copyright 2020 Marius van der Wijden
// This file is part of the fuzzy-vm library.
//
// The fuzzy-vm library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The fuzzy-vm library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the fuzzy-vm library. If not, see <http://www.gnu.org/licenses/>.

package generator

import (
	"github.com/MariusVanDerWijden/FuzzyVM/filler"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/bn256"
	"github.com/holiman/goevmlab/program"
)

var bn256addAddr = common.HexToAddress("0x6")

type bn256Caller struct{}

func (*bn256Caller) call(p *program.Program, f *filler.Filler) error {
	k := f.BigInt()
	point := new(bn256.G1).ScalarBaseMult(k)
	k2 := f.BigInt()
	point2 := new(bn256.G1).ScalarBaseMult(k2)
	c := callObj{
		gas:       f.BigInt(),
		address:   bn256addAddr,
		inOffset:  0,
		inSize:    128,
		outOffset: 0,
		outSize:   64,
		value:     f.BigInt(),
	}
	p.Mstore(point.Marshal(), 0)
	p.Mstore(point2.Marshal(), 64)
	callRandomizer(p, f, c)
	return nil
}
