import { readFileSync, mkdirSync, writeFileSync, existsSync, symlinkSync } from 'fs';
import { resolve, dirname } from 'path';
import type { Plugin } from 'vite';

export function extractIcons(): Plugin {
  return {
    name: 'extract-mdi-icons',
    configResolved(config) {
      const iconsJsonPath = resolve(
        dirname(import.meta.filename),
        'node_modules/@iconify-json/mdi/icons.json'
      );
      const outDir = resolve(config.publicDir, 'icons/mdi');

      if (existsSync(outDir)) return; // already extracted

      const data = JSON.parse(readFileSync(iconsJsonPath, 'utf-8'));
      const icons = data.icons as Record<string, { body: string }>;
      const aliases = data.aliases as Record<string, { parent: string }>;
      const width = data.width ?? 24;
      const height = data.height ?? 24;

      mkdirSync(outDir, { recursive: true });

      for (const [name, icon] of Object.entries(icons)) {
        const svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${width} ${height}">${icon.body}</svg>`;
        writeFileSync(resolve(outDir, `${name}.svg`), svg);
      }

      for (const [name, alias] of Object.entries(aliases)) {
        const target = `${alias.parent}.svg`;
        const link = resolve(outDir, `${name}.svg`);
        symlinkSync(target, link);
      }
    },
  };
}
