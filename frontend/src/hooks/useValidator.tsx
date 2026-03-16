import Ajv2020 from 'ajv/dist/2020';
import addFormats from 'ajv-formats';
import type { SchemaEntry } from '../api/resources';
import { useMemo } from 'react';

const ajv = new Ajv2020({ allErrors: true, strictSchema: false });
addFormats(ajv);

export const useValidator = (schema: SchemaEntry['schema']) => {
  const validator = useMemo(() => {
    if (!schema) {
      return null;
    }
    return ajv.compile(schema);
  }, [schema]);

  const validate = (obj: Record<string, unknown>): string | null => {
    if (!validator) return 'Schema not loaded';
    const valid = validator(obj);
    if (valid) {
      // no error
      return null;
    }
    const errors = validator.errors;
    if (!errors || errors.length === 0) {
      return 'Validator Error';
    }
    return errors
      .map((e) => {
        const field = e.instancePath || e.params?.missingProperty || '';
        const msg = e.message ?? 'invalid';
        return field ? `${field}: ${msg}` : msg;
      })
      .join('; ');
  };

  return validate;
};
