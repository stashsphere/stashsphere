import React from 'react';
import { SchemaCollection } from '../api/resources';

export const PropertySchemaCollectionContext = React.createContext<SchemaCollection>({
  aliases: {},
  schemas: {},
});
