#pragma once

#include <sys/types.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

#ifndef LIBQUANT_EXPORT
#ifdef _MSC_VER
#define LIBQUANT_EXPORT declspec(dllimport)
#else
#define LIBQUANT_EXPORT
#endif
#endif

typedef struct QRecord {
	uint16_t channelValue[2];
} QRecord;

typedef enum {
	QT_UNKNOWN = 0,
	QT_INT,
	QT_STRING,
} QVariantType;

typedef struct QVariant {
	QVariantType type;
	union {
		uintptr_t i;
		char* s;
	};
} QVariant;

typedef struct __QFile QFile;

LIBQUANT_EXPORT QFile* qf_open(const char* path);
LIBQUANT_EXPORT const float* qf_get_compensation_matrix(QFile* qfile, size_t idx);
LIBQUANT_EXPORT ssize_t qf_read_records(QFile* qfile, QRecord* records, size_t count);
LIBQUANT_EXPORT void qf_close(QFile* qfile);

#ifdef __cplusplus
}
#endif

// vim: set ts=8 sw=8 et ff=dos
