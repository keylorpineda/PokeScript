// Estándar de commits de PokeScript. Ver CONTRIBUTING.md.
// Formato: PKS type(scope): description   (mensajes en inglés)
export default {
  extends: ['@commitlint/config-conventional'],
  parserPreset: {
    parserOpts: {
      headerPattern: /^PKS (\w+)(?:\(([\w-]+)\))?!?: (.+)$/,
      headerCorrespondence: ['type', 'scope', 'subject'],
    },
  },
  plugins: [
    {
      rules: {
        'pks-prefix': ({ header }) => [
          /^PKS /.test(header ?? ''),
          'el encabezado debe iniciar con "PKS ", ej: PKS feat(lexer): recognize SINO_SI token',
        ],
      },
    },
  ],
  rules: {
    'pks-prefix': [2, 'always'],
    'type-enum': [
      2,
      'always',
      [
        'feat',
        'fix',
        'docs',
        'refactor',
        'test',
        'style',
        'perf',
        'build',
        'ci',
        'chore',
        'revert',
      ],
    ],
    'scope-enum': [
      2,
      'always',
      [
        'lexer',
        'parser',
        'analyzer',
        'interpreter',
        'assistant',
        'diagnostics',
        'project',
        'editor',
        'ui',
        'wails',
        'examples',
        'tests',
        'docs',
        'config',
      ],
    ],
    'subject-full-stop': [2, 'never', '.'],
    'header-max-length': [2, 'always', 100],
  },
};
