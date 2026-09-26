export const paths = {
  public: {
    home: '/',
    models: '/models',
    pricing: '/pricing',
    docs: '/docs',
    status: '/status',
    keyUsage: '/key-usage',
    imageStudio: '/image-studio',
    accountPortal: '/account-portal',
  },
  auth: {
    root: '/auth',
  },
  console: {
    root: '/console',
  },
  admin: {
    root: '/admin',
    gateway: {
      root: '/admin/gateway',
      accounts: '/admin/gateway/accounts',
      accountsGrok: '/admin/gateway/accounts/grok',
      accountsAntigravity: '/admin/gateway/accounts/antigravity',
      accountsClaude: '/admin/gateway/accounts/claude',
      accountsInvite: '/admin/gateway/accounts/invite',
      apiKeys: '/admin/gateway/api-keys',
      proxies: '/admin/gateway/proxies',
      images: '/admin/gateway/images',
      imageStudio: '/admin/gateway/images/studio',
      qualityTest: '/admin/gateway/quality-test',
      promptFilter: '/admin/gateway/prompt-filter',
      promptFilterOverview: '/admin/gateway/prompt-filter/overview',
      operations: '/admin/gateway/ops',
      operationsOverview: '/admin/gateway/ops/overview',
      runtime: '/admin/gateway/ops/runtime',
      errors: '/admin/gateway/ops/errors',
      scheduler: '/admin/gateway/ops/scheduler',
      usage: '/admin/gateway/usage',
      modelPricing: '/admin/gateway/model-pricing',
      payloadRules: '/admin/gateway/payload-rules',
      payloadRulesEditor: '/admin/gateway/payload-rules/editor',
      theme: '/admin/gateway/theme',
      settings: '/admin/gateway/settings',
      docs: '/admin/gateway/docs',
    },
  },
} as const

export function adminGatewayPath(path = '') {
  const suffix = path.startsWith('/') ? path : `/${path}`
  return path ? `${paths.admin.gateway.root}${suffix}` : paths.admin.gateway.root
}
