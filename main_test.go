package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)


// groupCount, skusPerGroup 	= 10, 2
// BenchmarkGenProducts-4		50000		      33 830 ns/op
// BenchmarkSqlxSingle-4		10			 169 663 172 ns/op
// BenchmarkSqlxValues-4		100		 	  10 087 356 ns/op

// groupCount, skusPerGroup     = 100, 2
// BenchmarkGenProducts-4		5000		     363 397 ns/op
// BenchmarkSqlxSingle-4		1		   1 699 034 966 ns/op
// BenchmarkSqlxValues-4   	    50			  40 177 339 ns/op

// groupCount, skusPerGroup     = 1000, 2
// BenchmarkGenProducts-4		500		       3 337 991 ns/op
// BenchmarkSqlxSingle-4   	    1		  14 730 843 327 ns/op
// BenchmarkSqlxValues-4        3			 347 748 834 ns/op
// BenchmarkGormSingle-4		1		  26 251 608 520 ns/op


func BenchmarkGenProducts(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenProducts(1000, 2, i)
	}
}

func BenchmarkSqlxSingle(b *testing.B) {
	s, err := Setup()
	require.NoError(b, err)

	db, err := s.DB()
	require.NoError(b, err)

	for i := 0; i < b.N; i++ {
		prods := GenProducts(1000, 2, i)
		_, err = SqlxSingle(db, prods)
		require.NoError(b, err)
	}
}

func BenchmarkSqlxValues(b *testing.B) {
	s, err := Setup()
	require.NoError(b, err)

	db, err := s.DB()
	require.NoError(b, err)

	for i := 0; i < b.N; i++ {
		prods := GenProducts(1000, 2, i)
		_, err = SqlxValues(db, prods)
		require.NoError(b, err)
	}
}

func BenchmarkGormSingle(b *testing.B) {
	s, err := Setup()
	require.NoError(b, err)

	db, err := s.GormDB()
	require.NoError(b, err)

	for i := 0; i < b.N; i++ {
		prods := GenProducts(1000, 2, i)
		_, err = GormSingle(db, prods)
		require.NoError(b, err)
	}
}
