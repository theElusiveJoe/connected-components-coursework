#include "mpi.h"
#include "stdlib.h"
#include <stdint.h>

static uint32_t* createArray(int size) {
	return (uint32_t*)malloc(sizeof(uint32_t) * size);
}

static void setArray(uint32_t *a, uint32_t s, int n) {
	a[n] = s;
}
static uint32_t getArray(uint32_t *a, int n) {
	return a[n];
}

static void freeArray(uint32_t *a) {
	free(a);
}