import { readFile, readdir, stat, writeFile } from 'node:fs/promises'
import { dirname, join, relative } from 'node:path'
import { fileURLToPath } from 'node:url'

const frontendRoot = join(dirname(fileURLToPath(import.meta.url)), '..')
const sourceRoot = join(frontendRoot, 'web', 'src')
const locales = ['zh', 'zh-TW', 'en']
const prefixes = ['common', 'accountPortal', 'apiKeys', 'imageStudioPortal', 'images', 'keyUsage', 'modelRequests', 'usage']
const dynamicSubtrees = ['apiKeys.status', 'keyUsage.range', 'imageStudioPortal.design.inspiration', 'imageStudioPortal.quota', 'images.status']
const publicCopy = {
  zh: {
    'imageStudioPortal.quota.quota_exhausted': '额度已用尽。已有作品仍可查看和下载。',
    'imageStudioPortal.quota.expired': '此 Key 已过期，请更换有效的 Key。',
    'imageStudioPortal.loginSubtitle': '粘贴你的 API Key，即可使用文生图与图生图。',
    'imageStudioPortal.loginHint': '仅用于本次服务鉴权，不会展示其他账号信息。',
    'accountPortal.submitSuccess': '已提交，安全审核通过后即可生效。感谢你的贡献！',
    'accountPortal.disabledDesc': '账号自助门户当前未开放。',
    'accountPortal.subheading': '用你自己的 ChatGPT 账号登录并授权。提交后将进行安全审核，全程不会向你索取或显示任何令牌。',
    'accountPortal.step1Desc': '用于在必要时联系你，不会作为登录账号。',
    'accountPortal.step3Desc': '确认已粘贴回跳地址后提交，等待安全审核。',
  },
  en: {
    'imageStudioPortal.quota.quota_exhausted': 'Quota exhausted. Existing images remain available to view and download.',
    'imageStudioPortal.quota.expired': 'This key has expired. Use a valid key to continue.',
    'imageStudioPortal.loginSubtitle': 'Paste your API key to generate and edit images.',
    'imageStudioPortal.loginHint': 'Used only to authorize this service and never to reveal other account information.',
    'accountPortal.submitSuccess': 'Submitted. It will become active after a security review. Thanks for contributing!',
    'accountPortal.disabledDesc': 'The account portal is not currently available.',
    'accountPortal.subheading': 'Sign in and authorize with your own ChatGPT account. Your submission goes through a security review, and we never ask for or display any token.',
    'accountPortal.step1Desc': 'Used to reach you if needed; it is not your login account.',
    'accountPortal.step3Desc': 'Paste the redirect URL, submit it, and wait for the security review.',
  },
}

async function filesUnder(directory) {
  const entries = await readdir(directory)
  const files = []
  for (const name of entries) {
    if (name === 'locales') continue
    const path = join(directory, name)
    const info = await stat(path)
    if (info.isDirectory()) files.push(...await filesUnder(path))
    else if (/\.(?:ts|tsx)$/.test(name)) files.push(path)
  }
  return files
}

function get(object, dottedPath) {
  return dottedPath.split('.').reduce((value, key) => value?.[key], object)
}

function set(object, dottedPath, value) {
  const parts = dottedPath.split('.')
  let cursor = object
  for (const part of parts.slice(0, -1)) cursor = cursor[part] ??= {}
  cursor[parts.at(-1)] = value
}

const referenced = new Set()
for (const path of await filesUnder(sourceRoot)) {
  const source = await readFile(path, 'utf8')
  const pattern = /['"`]([A-Za-z][A-Za-z0-9_-]*(?:\.[A-Za-z0-9_-]+)+)['"`]/g
  for (const match of source.matchAll(pattern)) {
    if (prefixes.some((prefix) => match[1].startsWith(prefix + '.'))) referenced.add(match[1])
  }
}

for (const language of locales) {
  const sourcePath = join(frontendRoot, 'admin', 'src', 'locales', `${language}.json`)
  const catalog = JSON.parse(await readFile(sourcePath, 'utf8'))
  const output = {}
  for (const key of referenced) {
    const value = get(catalog, key)
    if (value !== undefined) set(output, key, value)
  }
  for (const key of dynamicSubtrees) {
    const value = get(catalog, key)
    if (value !== undefined) set(output, key, value)
  }
  for (const [key, value] of Object.entries(publicCopy[language] ?? {})) set(output, key, value)
  set(output, 'settings.pricing.imageBilling.perImage', language === 'zh' ? '按成功图片张数' : language === 'zh-TW' ? '按成功圖片張數' : 'Per successful image')
  const outputPath = join(sourceRoot, 'locales', `${language}.public.json`)
  await writeFile(outputPath, JSON.stringify(output, null, 2) + '\n')
  console.log(`${relative(frontendRoot, outputPath)}: ${referenced.size} referenced keys`)
}
