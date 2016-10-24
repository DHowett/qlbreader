### libquant - a library for reading qlp/qlr/qlb files

#### EXAMPLE

```c
#include "quant.h"

int main() {
	QFile* qf = qf_open("A07.qlb");
	const float* comp1 = qf_get_compensation_matrix(qf, 0);

	QRecord rec = {0};
	while(qf_read_records(qf, &rec, 1) != 0) {
		// print record?
	}

	qf_close(qf);

	return 0;
}
```
