<script setup>
import { computed, nextTick, onErrorCaptured, onMounted, ref, watch } from 'vue'
import {
  GetAllTags,
  GetBillDetail,
  GetBills,
  GetCategoryTree,
  GetConfig,
  GetConfigStatus,
  UpdateConfig,
  GetDashboardStats,
  GetVerify,
  RunGitHubBackup,
  CreateBill,
  UpdateBillBasic,
  UpdateBillCategory,
  UpdateBillNote,
  UpdateBillTags,
  SoftDeleteBill,
  RestoreBill,
  GetDeletedBills,
  RunLocalBackup,
  RunReport,
  ExportData,
  OpenPath,
} from '../wailsjs/go/main/App'

function resolveDefaultMonth() {
  return new Date().toISOString().slice(0, 7)
}
const defaultMonth = resolveDefaultMonth()
const tagGroupLabels = {
  channel: '渠道',
  content: '内容',
  scenario: '场景',
  meal_time: '餐次',
  utility: '生活缴费',
  nature: '消费性质',
  income_related: '收入相关',
}
const topTagNames = [
  '手机支付',
  '淘宝',
  '淘宝闪购',
  '盒马',
  '拼多多',
  '线下店',
  '食材',
  '咖啡',
  '奶茶',
  '礼物',
  '节日',
  '午餐',
]

const commonMerchants = [
  '盒马',
  '淘宝',
  '淘宝闪购',
  '拼多多',
  '京东',
  '山姆',
  '线下店',
  '美团',
  '闲鱼',
  '手机支付',
]

const billTypes = [
  { value: 'expense', label: '支出' },
  { value: 'income', label: '收入' },
  { value: 'refund', label: '退款' },
  { value: 'reimbursement', label: '报销' },
  { value: 'transfer', label: '转账' },
  { value: 'adjustment', label: '调整' },
]

const month = ref(defaultMonth)
const verifyOutput = ref('')
const backupOutput = ref('')
const dashboardStats = ref(null)
const bills = ref([])
const categoryTree = ref([])
const allTags = ref([])
const selectedBillId = ref(null)
const billDetail = ref(null)
const detailForm = ref({ category: '', subCategory: '', tags: [], note: '' })
const originalDetail = ref(null)
const tagSearch = ref('')
const tagSearchActive = ref(false)
const searchQuery = ref('')
const filterType = ref('all')   // 'all' | 'expense' | 'income'
const filterCategory = ref('')  // category name or ''
const filterTag = ref('')       // tag name or ''

function onTagSearchFocus() { tagSearchActive.value = true }
function onTagSearchBlur() { setTimeout(() => { tagSearchActive.value = false }, 200) }
const showDeleted = ref(false)
const confirmDeleteId = ref(null)
const deletedBills = ref([])
const createMode = ref(false)
const detailPanelStatus = ref('closed') // 'closed' | 'loading' | 'edit' | 'create'
const errorMessage = ref('')
const detailMessage = ref('')
const loading = ref('')
const lastCreatedBill = ref(null)  // { billType, category, subCategory, billTime } from last create
const amountInputRef = ref(null)

const appConfig = ref(null)
const configError = ref('')
const configStatus = ref(null)

const showSettings = ref(false)
const showSettingsPanel = ref(false)
const settingsView = ref('main') // 'main' | 'deleted-bills'
const settingsForm = ref({ account_book_exe: '', db_path: '', backup_repo: '', default_month: '' })
const settingsError = ref('')
const settingsMessage = ref('')
const deletedBillsList = ref([])
const deletedBillsLoading = ref(false)
const deletedBillsError = ref('')
const restoringBillId = ref(null)
const maintenanceLoading = ref('') // '' | 'backup' | 'export' | 'report'
const maintenanceResult = ref('')
const maintenanceError = ref('')
const recentOperations = ref({
  sync: null,    // last sync time string
  backup: null,
  export: null,
  report: null,
})

async function loadConfig() {
  configError.value = ''
  try {
    const cfg = await GetConfig()
    appConfig.value = cfg
    if (cfg.default_month && cfg.default_month !== 'current') {
      month.value = cfg.default_month
      await refreshDashboard()
    }
  } catch (error) {
    configError.value = error?.message || String(error)
    appConfig.value = null
  }
  try {
    configStatus.value = await GetConfigStatus()
  } catch (error) {
    // GetConfigStatus always succeeds; fallback
    configStatus.value = null
  }
}

function openSettings() {
  settingsError.value = ''
  settingsMessage.value = ''
  settingsView.value = 'main'
  if (appConfig.value) {
    settingsForm.value = {
      account_book_exe: appConfig.value.account_book_exe || '',
      db_path: appConfig.value.db_path || '',
      backup_repo: appConfig.value.backup_repo || '',
      default_month: appConfig.value.default_month || 'current',
    }
  }
  showSettings.value = true
  showSettingsPanel.value = true
}

function closeSettings() {
  showSettings.value = false
  showSettingsPanel.value = false
  settingsView.value = 'main'
  settingsError.value = ''
  settingsMessage.value = ''
}

async function saveSettings() {
  settingsError.value = ''
  settingsMessage.value = ''
  loading.value = 'settings'
  try {
    await UpdateConfig(settingsForm.value)
    settingsMessage.value = '配置已保存'
    await loadConfig()
    if (configStatus.value && configStatus.value.overall_status === 'ok') {
      await Promise.all([refreshDashboard(), refreshBills()])
    }
  } catch (error) {
    settingsError.value = error?.message || String(error)
    await loadConfig()
  } finally {
    loading.value = ''
  }
}

async function openDeletedBillsView() {
  deletedBillsError.value = ''
  deletedBillsLoading.value = true
  deletedBillsList.value = []
  settingsView.value = 'deleted-bills'
  try {
    deletedBillsList.value = (await GetDeletedBills(month.value)) || []
  } catch (error) {
    deletedBillsError.value = error?.message || String(error)
  } finally {
    deletedBillsLoading.value = false
  }
}

function goBackToSettings() {
  settingsView.value = 'main'
}

async function restoreDeletedBill(id) {
  restoringBillId.value = id
  deletedBillsError.value = ''
  try {
    await RestoreBill(id)
    deletedBillsList.value = deletedBillsList.value.filter(b => b.id !== id)
    await Promise.all([refreshBills(), refreshDashboard()])
  } catch (error) {
    deletedBillsError.value = error?.message || String(error)
  } finally {
    restoringBillId.value = null
  }
}

async function runMaintenanceBackup() {
  maintenanceLoading.value = 'backup'
  maintenanceResult.value = ''
  maintenanceError.value = ''
  try {
    const output = await RunLocalBackup(month.value)
    maintenanceResult.value = output
    recentOperations.value.backup = new Date().toISOString()
  } catch (error) {
    maintenanceError.value = error?.message || String(error)
  } finally {
    maintenanceLoading.value = ''
  }
}

async function runMaintenanceExport() {
  maintenanceLoading.value = 'export'
  maintenanceResult.value = ''
  maintenanceError.value = ''
  try {
    const output = await ExportData(month.value)
    maintenanceResult.value = output
    recentOperations.value.export = new Date().toISOString()
  } catch (error) {
    maintenanceError.value = error?.message || String(error)
  } finally {
    maintenanceLoading.value = ''
  }
}

async function runMaintenanceReport() {
  maintenanceLoading.value = 'report'
  maintenanceResult.value = ''
  maintenanceError.value = ''
  try {
    const output = await RunReport(month.value)
    maintenanceResult.value = output
    recentOperations.value.report = new Date().toISOString()
  } catch (error) {
    maintenanceError.value = error?.message || String(error)
  } finally {
    maintenanceLoading.value = ''
  }
}

function getMaintenanceDir(kind) {
  if (!appConfig.value) return ''
  if (kind === 'data') {
    const db = appConfig.value.db_path || ''
    const sep = db.lastIndexOf('\\')
    return sep > 0 ? db.substring(0, sep) : db
  }
  const repo = appConfig.value.backup_repo || ''
  return repo ? repo + '\\' + kind : ''
}

async function handleOpenPath(path) {
  maintenanceError.value = ''
  try {
    await OpenPath(path)
  } catch (error) {
    maintenanceError.value = error?.message || String(error)
  }
}

function openDataDir() { handleOpenPath(getMaintenanceDir('data')) }
function openBackupDir() { handleOpenPath(getMaintenanceDir('backups')) }
function openExportDir() { handleOpenPath(getMaintenanceDir('exports')) }
function openReportDir() { handleOpenPath(getMaintenanceDir('reports')) }

function onDetailPanelKeydown(event) {
  if (showSettingsPanel.value) return
  if (detailPanelStatus.value === 'closed') return

  if (event.key === 'Escape') {
    event.preventDefault()
    closeDetailPanel()
    return
  }

  if (createMode.value) {
    const target = event.target
    const isTextarea = target?.tagName === 'TEXTAREA'

    if (event.key === 'Enter' && event.ctrlKey) {
      event.preventDefault()
      saveBillDetail(true)
      return
    }

    if (event.key === 'Enter' && !isTextarea) {
      event.preventDefault()
      saveBillDetail(false)
      return
    }
  }
}

function registerSyncTime() {
  recentOperations.value.sync = new Date().toISOString()
}

onErrorCaptured((err, _instance, info) => {
  console.error('[Vue error]', err?.message || err, info)
  errorMessage.value = `渲染错误: ${err?.message || err}`
  return false
})

const moneyCards = computed(() => [
  { label: '收入', value: formatMoney(dashboardStats.value?.income) },
  { label: '支出', value: formatMoney(dashboardStats.value?.expense) },
  { label: '结余', value: formatMoney(dashboardStats.value?.balance) },
])

const statusCards = computed(() => [
  { label: '有效账单', value: formatNumber(dashboardStats.value?.active_bills) },
  { label: '已作废', value: formatNumber(dashboardStats.value?.deleted_bills) },
  { label: '最近同步时间', value: formatDateTime(dashboardStats.value?.last_backup_time), compact: true },
])

const billGroups = computed(() => {
  const groups = []
  const index = new Map()
  for (const bill of displayBills.value) {
    const dateKey = bill.bill_time.slice(0, 10)
    if (!index.has(dateKey)) {
      const group = { date: dateKey, label: formatLedgerDate(dateKey), bills: [], expense: 0, income: 0 }
      index.set(dateKey, group)
      groups.push(group)
    }
    const group = index.get(dateKey)
    if (bill.type === 'expense') group.expense += bill.amount
    else if (bill.type === 'income') group.income += bill.amount
    group.bills.push(bill)
  }
  return groups
})

const expenseBills = computed(() => bills.value.filter(b => b.type === 'expense'))

const categoryTop5 = computed(() => {
  const map = new Map()
  for (const bill of expenseBills.value) {
    const cat = bill.category || '未分类'
    const prev = map.get(cat) || { category: cat, total: 0, count: 0 }
    prev.total += bill.amount
    prev.count++
    map.set(cat, prev)
  }
  const totalExpense = expenseBills.value.reduce((s, b) => s + b.amount, 0)
  return [...map.values()]
    .sort((a, b) => b.total - a.total)
    .slice(0, 5)
    .map(item => ({ ...item, pct: totalExpense > 0 ? (item.total / totalExpense * 100) : 0 }))
})

const merchantTop5 = computed(() => {
  const map = new Map()
  for (const bill of expenseBills.value) {
    const m = bill.merchant
    if (!m) continue
    const prev = map.get(m) || { merchant: m, total: 0, count: 0 }
    prev.total += bill.amount
    prev.count++
    map.set(m, prev)
  }
  return [...map.values()]
    .sort((a, b) => b.total - a.total)
    .slice(0, 5)
})

const largeExpenses = computed(() => {
  return [...expenseBills.value]
    .sort((a, b) => b.amount - a.amount)
    .slice(0, 5)
})

const dailyStats = computed(() => {
  const now = new Date()
  const [y, m] = month.value.split('-').map(Number)
  const isCurrentMonth = y === now.getFullYear() && m === now.getMonth() + 1
  const daysInMonth = new Date(y, m, 0).getDate()
  const elapsedDays = isCurrentMonth ? now.getDate() : daysInMonth
  const remainingDays = Math.max(0, daysInMonth - elapsedDays)
  const totalExpense = expenseBills.value.reduce((s, b) => s + b.amount, 0)
  const dailyAvg = elapsedDays > 0 ? totalExpense / elapsedDays : 0
  return { dailyAvg, elapsedDays, remainingDays, daysInMonth, isCurrentMonth, totalExpense }
})

const categoryType = computed(() => (detailForm.value.billType === 'income' ? 'income' : 'expense'))
const availableCategories = computed(() =>
  categoryTree.value.filter((category) => category.type === categoryType.value),
)
const orderedCategories = computed(() => moveSelectedFirst(availableCategories.value, detailForm.value.category))
const selectedCategory = computed(() =>
  availableCategories.value.find((category) => category.name === detailForm.value.category),
)
const availableSubCategories = computed(() => selectedCategory.value?.children || [])
const orderedSubCategories = computed(() =>
  moveSelectedFirst(availableSubCategories.value, detailForm.value.subCategory),
)
const hasMissingCategory = computed(() => {
  if (!billDetail.value) {
    return false
  }
  return !availableCategories.value.some((category) => category.name === billDetail.value.category)
})

const selectedTagSet = computed(() => new Set(detailForm.value.tags))
const tagCandidateGroups = computed(() => {
  const query = tagSearch.value.trim().toLowerCase()
  const candidates = query ? searchedTags(query) : commonTags()
  return groupTags(candidates)
})

const filteredBills = computed(() => {
  let result = bills.value

  if (filterType.value !== 'all') {
    result = result.filter(b => b.type === filterType.value)
  }
  if (filterCategory.value) {
    result = result.filter(b => (b.category || '未分类') === filterCategory.value)
  }
  if (filterTag.value) {
    result = result.filter(b => Array.isArray(b.tags) && b.tags.includes(filterTag.value))
  }

  const query = searchQuery.value.trim().toLowerCase()
  if (query) {
    result = result.filter(bill => {
      const searchable = [
        bill.note, bill.merchant, bill.category, bill.sub_category,
        String(bill.amount || ''), bill.bill_time ? bill.bill_time.slice(0, 10) : '',
        ...(Array.isArray(bill.tags) ? bill.tags : [])
      ].filter(Boolean).join(' ').toLowerCase()
      return searchable.includes(query)
    })
  }

  return result
})

const displayBills = computed(() => {
  return showDeleted.value ? deletedBills.value : filteredBills.value
})

const availableFilterCategories = computed(() => {
  const cats = new Map()
  for (const bill of bills.value) {
    const c = bill.category || '未分类'
    if (!cats.has(c)) cats.set(c, true)
  }
  return [...cats.keys()].slice(0, 10)
})

const availableFilterTags = computed(() => {
  const tags = new Map()
  for (const bill of bills.value) {
    for (const tag of (Array.isArray(bill.tags) ? bill.tags : [])) {
      if (!tags.has(tag)) tags.set(tag, true)
    }
  }
  return [...tags.keys()].slice(0, 10)
})

const hasActiveFilters = computed(() =>
  filterType.value !== 'all' || filterCategory.value !== '' || filterTag.value !== '' || searchQuery.value.trim() !== ''
)

const filterResultText = computed(() => {
  if (!hasActiveFilters.value) return ''
  if (showDeleted.value) return ''
  const total = bills.value.length
  const filtered = filteredBills.value.length
  if (filtered === total) return `${total} 条`
  return `已筛选 ${filtered} / ${total} 条`
})

function clearFilters() {
  searchQuery.value = ''
  filterType.value = 'all'
  filterCategory.value = ''
  filterTag.value = ''
}

function toggleFilterCategory(cat) {
  filterCategory.value = filterCategory.value === cat ? '' : cat
}

function toggleFilterTag(tag) {
  filterTag.value = filterTag.value === tag ? '' : tag
}

const hasUnsavedChanges = computed(() => {
  if (!originalDetail.value) {
    return false
  }
  return (
    detailForm.value.amount !== originalDetail.value.amount ||
    detailForm.value.billTime !== originalDetail.value.billTime ||
    detailForm.value.billType !== originalDetail.value.billType ||
    detailForm.value.category !== originalDetail.value.category ||
    detailForm.value.subCategory !== originalDetail.value.subCategory ||
    detailForm.value.merchant !== originalDetail.value.merchant ||
    detailForm.value.note !== originalDetail.value.note ||
    tagKey(detailForm.value.tags) !== tagKey(originalDetail.value.tags)
  )
})

const recentCategoryPairs = computed(() => {
  const seen = new Set()
  const pairs = []
  for (const bill of bills.value) {
    if (!bill.category || !bill.sub_category) continue
    const key = bill.category + '\x1f' + bill.sub_category
    if (seen.has(key)) continue
    seen.add(key)
    pairs.push({ category: bill.category, subCategory: bill.sub_category })
    if (pairs.length >= 5) break
  }
  return pairs
})

const recentMerchants = computed(() => {
  const seen = new Set()
  const merchants = []
  for (const bill of bills.value) {
    const m = bill.merchant
    if (!m) continue
    if (seen.has(m)) continue
    seen.add(m)
    merchants.push(m)
    if (merchants.length >= 5) break
  }
  return merchants
})

function selectRecentCategoryPair(pair) {
  detailForm.value.category = pair.category
  detailForm.value.subCategory = pair.subCategory
}

watch(month, () => {
  if (detailPanelStatus.value !== 'closed') {
    closeDetailPanel()
  }
  showDeleted.value = false
  deletedBills.value = []
  searchQuery.value = ''
  filterType.value = 'all'
  filterCategory.value = ''
  filterTag.value = ''
  lastCreatedBill.value = null
})

watch([filteredBills, showDeleted], () => {
  if (selectedBillId.value && !displayBills.value.some(b => b.id === selectedBillId.value)) {
    closeDetailPanel()
  }
})

watch(() => detailForm.value.billType, () => {
  normalizeCategorySelection()
})

const footerStatus = computed(() => {
  if (detailPanelStatus.value === 'create') {
    return detailMessage.value || '未保存'
  }
  if (loading.value === 'save-detail') {
    return '保存中'
  }
  if (hasUnsavedChanges.value) {
    return '有未保存修改'
  }
  return detailMessage.value || '已保存'
})

onMounted(async () => {
  await Promise.all([loadConfig(), loadReferenceData(), refreshDashboard()])
})

async function loadReferenceData() {
  const [categories, tags] = await Promise.all([GetCategoryTree(), GetAllTags()])
  categoryTree.value = categories || []
  allTags.value = tags || []
}

async function refreshDashboard() {
  loading.value = 'dashboard'
  errorMessage.value = ''
  try {
    dashboardStats.value = await GetDashboardStats(month.value)
    bills.value = (await GetBills(month.value)) || []
  } catch (error) {
    errorMessage.value = error?.message || String(error)
  } finally {
    loading.value = ''
  }
}

async function refreshBills() {
  bills.value = await GetBills(month.value)
}

async function openBillDetail(id) {
  console.log('[openBillDetail] bill.id =', id, 'showDeleted =', showDeleted.value, 'createMode =', createMode.value)
  createMode.value = false
  detailPanelStatus.value = 'loading'
  selectedBillId.value = id
  detailMessage.value = ''
  loading.value = 'detail'
  try {
    if (categoryTree.value.length === 0 || allTags.value.length === 0) {
      await loadReferenceData()
    }
    const detail = await GetBillDetail(id)
    console.log('[openBillDetail] detail =', detail?.id, detail?.display_title, JSON.stringify(detail).slice(0, 120))
    setBillDetail(detail)
    detailPanelStatus.value = 'edit'
  } catch (error) {
    console.error('[openBillDetail] error =', error?.message || error)
    errorMessage.value = error?.message || String(error)
    detailPanelStatus.value = 'closed'
  } finally {
    loading.value = ''
  }
}

function formatDateTimeLocal(dateStr) {
  if (!dateStr) return ''
  return dateStr.replace(' ', 'T').slice(0, 16)
}

function formatBillTime(dateTimeLocalStr) {
  if (!dateTimeLocalStr) return ''
  return dateTimeLocalStr.replace('T', ' ') + ':00'
}

function validateForm() {
  if (!detailForm.value.amount || detailForm.value.amount <= 0) {
    detailMessage.value = '请输入有效金额'
    return false
  }
  if (!detailForm.value.billTime) {
    detailMessage.value = '请选择时间'
    return false
  }
  if (!detailForm.value.billType) {
    detailMessage.value = '请选择类型'
    return false
  }
  if (!detailForm.value.category) {
    detailMessage.value = '请选择分类'
    return false
  }
  if (!detailForm.value.subCategory) {
    detailMessage.value = '请选择子分类'
    return false
  }
  return true
}

function openCreateBill() {
  createMode.value = true
  detailPanelStatus.value = 'create'
  selectedBillId.value = null
  billDetail.value = null
  originalDetail.value = null
  detailMessage.value = ''
  tagSearch.value = ''

  const prev = lastCreatedBill.value
  const today = nowLocalDateTime()
  detailForm.value = {
    amount: null,
    billTime: prev?.billTime || today,
    billType: prev?.billType || 'expense',
    category: prev?.category || '',
    subCategory: prev?.subCategory || '',
    merchant: '',
    tags: [],
    note: '',
  }
  normalizeCategorySelection()
}

function nowLocalDateTime() {
  const now = new Date()
  const pad = (n) => String(n).padStart(2, '0')
  return now.getFullYear() + '-' + pad(now.getMonth() + 1) + '-' + pad(now.getDate()) + 'T' + pad(now.getHours()) + ':' + pad(now.getMinutes())
}

function setBillDetail(detail) {
  billDetail.value = detail
  detailForm.value = {
    amount: detail?.amount || null,
    billTime: formatDateTimeLocal(detail?.bill_time),
    billType: detail?.type || 'expense',
    category: detail?.category || '',
    subCategory: detail?.sub_category || '',
    merchant: detail?.merchant || '',
    tags: [...(detail?.tags || [])],
    note: detail?.note || '',
  }
  originalDetail.value = {
    amount: detailForm.value.amount,
    billTime: detailForm.value.billTime,
    billType: detailForm.value.billType,
    category: detailForm.value.category,
    subCategory: detailForm.value.subCategory,
    merchant: detailForm.value.merchant,
    tags: [...detailForm.value.tags],
    note: detailForm.value.note,
  }
  tagSearch.value = ''
  normalizeCategorySelection()
}

function closeDetailPanel() {
  detailPanelStatus.value = 'closed'
  selectedBillId.value = null
  billDetail.value = null
  originalDetail.value = null
  createMode.value = false
  detailMessage.value = ''
  confirmDeleteId.value = null
}

function selectCategory(categoryName) {
  detailForm.value.category = categoryName
  const children = availableSubCategories.value
  const hasCurrentSubCategory = children.some((child) => child.name === detailForm.value.subCategory)
  if (!hasCurrentSubCategory) {
    detailForm.value.subCategory = children[0]?.name || ''
  }
}

function selectSubCategory(subCategoryName) {
  detailForm.value.subCategory = subCategoryName
}

function normalizeCategorySelection() {
  const hasCategory = availableCategories.value.some(
    (category) => category.name === detailForm.value.category,
  )
  if (!hasCategory && availableCategories.value.length > 0) {
    selectCategory(availableCategories.value[0].name)
    return
  }
  selectCategory(detailForm.value.category)
}

function addTag(tagName) {
  if (!detailForm.value.tags.includes(tagName)) {
    detailForm.value.tags = [...detailForm.value.tags, tagName]
  }
  tagSearch.value = ''
}

function removeTag(tag) {
  detailForm.value.tags = detailForm.value.tags.filter((item) => item !== tag)
}

async function saveBillDetail(closeAfter = false) {
  if (detailPanelStatus.value === 'closed') {
    return
  }
  if (!validateForm()) {
    return
  }
  loading.value = 'save-detail'
  errorMessage.value = ''
  detailMessage.value = ''
  try {
    if (detailPanelStatus.value === 'create') {
      await CreateBill({
        bill_time: formatBillTime(detailForm.value.billTime),
        type: detailForm.value.billType,
        amount: detailForm.value.amount,
        category: detailForm.value.category,
        sub_category: detailForm.value.subCategory,
        merchant: detailForm.value.merchant,
        tags: detailForm.value.tags,
        note: detailForm.value.note,
      })
      lastCreatedBill.value = {
        billType: detailForm.value.billType,
        category: detailForm.value.category,
        subCategory: detailForm.value.subCategory,
        billTime: detailForm.value.billTime,
      }
      await Promise.all([refreshBills(), refreshDashboard()])

      if (closeAfter) {
        closeDetailPanel()
      } else {
        detailForm.value.amount = null
        detailForm.value.note = ''
        detailForm.value.tags = []
        detailForm.value.merchant = ''
        detailMessage.value = '已保存'
        nextTick(() => { amountInputRef.value?.focus() })
      }
    } else {
      const id = billDetail.value.id
      await UpdateBillBasic({
        id: id,
        bill_time: formatBillTime(detailForm.value.billTime),
        type: detailForm.value.billType,
        amount: detailForm.value.amount,
        merchant: detailForm.value.merchant,
      })
      await UpdateBillCategory(id, detailForm.value.category, detailForm.value.subCategory)
      await UpdateBillTags(id, detailForm.value.tags)
      await UpdateBillNote(id, detailForm.value.note)
      await Promise.all([refreshBills(), refreshDashboard()])
      setBillDetail(await GetBillDetail(id))
      detailMessage.value = '已保存'
    }
  } catch (error) {
    detailMessage.value = error?.message || String(error)
  } finally {
    loading.value = ''
  }
}

async function runVerify() {
  loading.value = 'verify'
  errorMessage.value = ''
  try {
    verifyOutput.value = await GetVerify()
    await refreshDashboard()
  } catch (error) {
    errorMessage.value = error?.message || String(error)
  } finally {
    loading.value = ''
  }
}

async function runBackup() {
  loading.value = 'backup'
  errorMessage.value = ''
  try {
    backupOutput.value = await RunGitHubBackup(month.value)
    registerSyncTime()
    await refreshDashboard()
  } catch (error) {
    errorMessage.value = error?.message || String(error)
  } finally {
    loading.value = ''
  }
}

async function toggleShowDeleted() {
  showDeleted.value = !showDeleted.value
  if (showDeleted.value) {
    loading.value = 'deleted'
    try {
      deletedBills.value = (await GetDeletedBills(month.value)) || []
    } catch (error) {
      errorMessage.value = error?.message || String(error)
    } finally {
      loading.value = ''
    }
  }
  closeDetailPanel()
}

async function softDeleteBill(id) {
  loading.value = 'delete'
  detailMessage.value = ''
  try {
    await SoftDeleteBill(id)
    closeDetailPanel()
    await Promise.all([refreshBills(), refreshDashboard()])
    confirmDeleteId.value = null
  } catch (error) {
    detailMessage.value = error?.message || String(error)
  } finally {
    loading.value = ''
  }
}

async function restoreBill(id) {
  loading.value = 'restore'
  detailMessage.value = ''
  try {
    await RestoreBill(id)
    closeDetailPanel()
    if (showDeleted.value) {
      await toggleShowDeleted()
    }
    await Promise.all([refreshBills(), refreshDashboard()])
  } catch (error) {
    detailMessage.value = error?.message || String(error)
  } finally {
    loading.value = ''
  }
}

function confirmDelete(id) {
  confirmDeleteId.value = id
}

function cancelDelete() {
  confirmDeleteId.value = null
}

function commonTags() {
  const byName = new Map(allTags.value.map((tag) => [tag.name, tag]))
  const result = []
  for (const name of topTagNames) {
    const tag = byName.get(name)
    if (tag && tag.group_name !== 'channel' && !selectedTagSet.value.has(tag.name)) {
      result.push(tag)
    }
    if (result.length >= 8) {
      return result
    }
  }
  for (const tag of allTags.value) {
    if (tag.group_name === 'channel') continue
    if (!selectedTagSet.value.has(tag.name) && !result.some((item) => item.name === tag.name)) {
      result.push(tag)
    }
    if (result.length >= 8) {
      break
    }
  }
  return result
}

function searchedTags(query) {
  const result = []
  for (const tag of allTags.value) {
    if (tag.group_name === 'channel') continue
    if (selectedTagSet.value.has(tag.name)) {
      continue
    }
    const searchable = `${tag.name} ${tag.group_name} ${groupLabel(tag.group_name)}`.toLowerCase()
    if (searchable.includes(query)) {
      result.push(tag)
    }
    if (result.length >= 24) {
      break
    }
  }
  return result
}

function groupTags(tags) {
  const grouped = []
  const groupIndex = new Map()
  for (const tag of tags) {
    if (!groupIndex.has(tag.group_name)) {
      const group = { name: tag.group_name, label: groupLabel(tag.group_name), tags: [] }
      groupIndex.set(tag.group_name, group)
      grouped.push(group)
    }
    groupIndex.get(tag.group_name).tags.push(tag)
  }
  return grouped
}

function moveSelectedFirst(items, selectedName) {
  if (!selectedName) {
    return items
  }
  const selected = items.find((item) => item.name === selectedName)
  if (!selected) {
    return items
  }
  return [selected, ...items.filter((item) => item.name !== selectedName)]
}

function formatMoney(value) {
  if (typeof value !== 'number' || Number.isNaN(value)) {
    return '--'
  }
  return new Intl.NumberFormat('zh-CN', {
    style: 'currency',
    currency: 'CNY',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(value)
}

function formatNumber(value) {
  if (typeof value !== 'number' || Number.isNaN(value)) {
    return '--'
  }
  return new Intl.NumberFormat('zh-CN').format(value)
}

function formatDateTime(value) {
  if (!value) {
    return '暂无记录'
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return '暂无记录'
  }
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

function formatLedgerDate(dateKey) {
  const date = new Date(`${dateKey}T00:00:00`)
  if (Number.isNaN(date.getTime())) {
    return dateKey
  }
  const weekdays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
  return `${dateKey.slice(5, 7)}-${dateKey.slice(8, 10)} ${weekdays[date.getDay()]}`
}

function billTitle(bill) {
  return bill.display_title || bill.merchant || bill.note || '未命名账单'
}

function billSubtitle(bill) {
  return bill.display_subtitle || ''
}

function listTags(bill) {
  const tags = bill?.tags
  if (!Array.isArray(tags) || tags.length === 0) return []
  return tags.filter(tag => tag !== bill.merchant)
}

function categoryText(bill) {
  const parts = [bill.category, bill.sub_category].filter(Boolean)
  return parts.length > 0 ? parts.join(' / ') : '未分类'
}

function typeLabel(type) {
  const labels = {
    expense: '支出',
    income: '收入',
    refund: '退款',
    reimbursement: '报销',
    transfer: '转账',
    adjustment: '调整',
  }
  return labels[type] || type
}

function amountClass(type) {
  return type === 'income' ? 'income-amount' : 'expense-amount'
}

function formatBillAmount(bill) {
  if (!bill || typeof bill.amount !== 'number') return '--'
  const formatted = new Intl.NumberFormat('zh-CN', {
    style: 'currency',
    currency: 'CNY',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(Math.abs(bill.amount))
  return bill.type === 'income' ? '+' + formatted : '-' + formatted
}

function groupLabel(groupName) {
  return tagGroupLabels[groupName] || groupName
}

function tagKey(tags) {
  return [...tags].sort((a, b) => a.localeCompare(b, 'zh-CN')).join('\x1f')
}
</script>

<template>
  <div class="app-shell">
    <!-- Sidebar -->
    <aside class="sidebar">
      <div class="sidebar-brand">
        <h2>MiraLedger</h2>
        <p>Personal finance</p>
      </div>
      <nav class="sidebar-nav">
        <span class="sidebar-section-label">导航</span>
        <button class="sidebar-item active">
          <span class="sidebar-icon">■</span> 概览
        </button>
        <button class="sidebar-item">
          <span class="sidebar-icon">≡</span> 账单流
        </button>
        <button class="sidebar-item">
          <span class="sidebar-icon">⚙</span> 设置
        </button>
      </nav>
      <div class="sidebar-footer">
        <span class="sidebar-section-label">状态</span>
        <div class="sidebar-status-card">
          <span
            v-if="configStatus"
            class="config-status-inline"
            :class="{ 'config-err': configStatus.overall_status === 'error' }"
            style="display:block;padding:0"
          >
            {{ configStatus.overall_status === 'error' ? '配置异常' : '配置正常' }}
          </span>
        </div>
      </div>
    </aside>

    <!-- Main Area -->
    <div class="main-area">
    <main class="dashboard" :class="{ 'has-detail-panel': detailPanelStatus !== 'closed' || showSettingsPanel }">
    <header class="topbar">
      <div class="topbar-left">
        <p class="eyebrow">Overview</p>
        <h1>概览</h1>
      </div>

      <div class="actions">
        <label class="month-picker">
          {{ month }}
          <input v-model="month" type="month" @change="refreshDashboard" />
        </label>
        <button type="button" class="btn-secondary" :disabled="loading !== ''" @click="refreshDashboard">
          {{ loading === 'dashboard' ? '刷新中' : '刷新' }}
        </button>
        <button type="button" class="btn-secondary" :disabled="loading !== ''" @click="runBackup">
          {{ loading === 'backup' ? '同步中' : '同步' }}
        </button>
        <button v-if="configStatus" type="button" class="btn-ghost" @click="openSettings">设置</button>
      </div>
    </header>

    <p v-if="errorMessage" class="notice">{{ errorMessage }}</p>

    <div v-if="configStatus?.overall_status === 'error'" class="config-error-banner">
      <span>配置异常：{{ configStatus.errors[0] || '请检查配置' }}</span>
      <button type="button" @click="openSettings">打开设置</button>
    </div>

    <div class="section-divider"><span class="section-divider-label">概览</span></div>

    <section class="card-grid primary-grid" aria-label="月度财务概览">
      <article v-for="card in moneyCards" :key="card.label" class="stat-card">
        <span>{{ card.label }}</span>
        <strong>{{ card.value }}</strong>
      </article>
    </section>

    <section class="card-grid secondary-grid" aria-label="账本状态">
      <article
        v-for="card in statusCards"
        :key="card.label"
        class="stat-card muted-card"
        :class="{ compact: card.compact }"
      >
        <span>{{ card.label }}</span>
        <strong>{{ card.value }}</strong>
      </article>
    </section>

    <div class="section-divider"><span class="section-divider-label">统计</span></div>

    <section class="stats-grid" aria-label="月度统计">
      <!-- 本月花在哪 -->
      <article class="stats-module">
        <h3 class="stats-module-title">本月花在哪</h3>
        <div v-if="categoryTop5.length > 0" class="stats-list">
          <div v-for="item in categoryTop5" :key="item.category" class="stats-row stats-row-cat">
            <span class="stats-label">{{ item.category }}</span>
            <span class="stats-value">{{ formatMoney(item.total) }}</span>
            <span class="stats-pct">{{ Math.round(item.pct) }}%</span>
          </div>
        </div>
        <p v-else class="stats-empty">本月暂无支出</p>
      </article>

      <!-- 常去商家 -->
      <article class="stats-module">
        <h3 class="stats-module-title">常去商家</h3>
        <div v-if="merchantTop5.length > 0" class="stats-list">
          <div v-for="item in merchantTop5" :key="item.merchant" class="stats-row">
            <span class="stats-label">{{ item.merchant }}</span>
            <span class="stats-value">{{ formatMoney(item.total) }}</span>
            <span class="stats-count">{{ item.count }} 笔</span>
          </div>
        </div>
        <p v-else class="stats-empty">暂无商家记录</p>
      </article>

      <!-- 本月大额支出 -->
      <article class="stats-module">
        <h3 class="stats-module-title">本月大额支出</h3>
        <div v-if="largeExpenses.length > 0" class="stats-list">
          <div
            v-for="bill in largeExpenses"
            :key="'le-'+bill.id"
            class="stats-row stats-row-clickable"
            @click="openBillDetail(bill.id)"
          >
            <span class="stats-label">
              <span class="stats-date">{{ bill.bill_time.slice(5, 10) }}</span>
              {{ categoryText(bill) }}
            </span>
            <span class="stats-sub" v-if="bill.merchant">{{ bill.merchant }}</span>
            <span class="stats-value">{{ formatBillAmount(bill) }}</span>
          </div>
        </div>
        <p v-else class="stats-empty">本月暂无大额支出</p>
      </article>

      <!-- 本月概况 -->
      <article class="stats-module">
        <h3 class="stats-module-title">本月概况</h3>
        <div class="stats-list">
          <div class="stats-row">
            <span class="stats-label">日均支出</span>
            <span class="stats-value">{{ formatMoney(dailyStats.dailyAvg) }}</span>
          </div>
          <div class="stats-row">
            <span class="stats-label">总支出</span>
            <span class="stats-value">{{ formatMoney(dailyStats.totalExpense) }}</span>
          </div>
          <div class="stats-row">
            <span class="stats-label">支出笔数</span>
            <span class="stats-value">{{ expenseBills.length }} 笔</span>
          </div>
          <div v-if="dailyStats.isCurrentMonth" class="stats-row">
            <span class="stats-label">已过天数</span>
            <span class="stats-value">{{ dailyStats.elapsedDays }} 天 / {{ dailyStats.daysInMonth }} 天</span>
          </div>
          <div v-if="dailyStats.isCurrentMonth && dailyStats.remainingDays > 0" class="stats-row">
            <span class="stats-label">剩余天数</span>
            <span class="stats-value">{{ dailyStats.remainingDays }} 天</span>
          </div>
        </div>
      </article>
    </section>

    <div class="section-divider"><span class="section-divider-label">账单</span></div>

    <section class="ledger-section" aria-label="账单流">
      <div class="ledger-sticky-header">
        <div class="ledger-toolbar">
          <div class="ledger-toolbar-left">
            <p class="eyebrow">Ledger</p>
            <h2>账单流</h2>
          </div>
          <div class="ledger-toolbar-right">
            <span v-if="filterResultText" class="ledger-count">{{ filterResultText }}</span>
            <span v-else class="ledger-count">{{ filteredBills.length }} 条</span>
            <button type="button" class="btn-create" :disabled="loading !== ''" @click="openCreateBill">
              + 记一笔
            </button>
          </div>
        </div>
        <div class="ledger-toolbar-row">
          <input v-model="searchQuery" type="search" placeholder="搜索分类、商家、备注、标签..." class="bill-search" />
          <button type="button" class="btn-toggle-deleted" :class="{ active: showDeleted }" @click="toggleShowDeleted">
            {{ showDeleted ? '返回账单' : '已作废' }}
          </button>
        </div>
        <div v-if="!showDeleted" class="ledger-filter-row">
          <div class="filter-chips">
            <button
              v-for="opt in [{v:'all',l:'全部'},{v:'expense',l:'支出'},{v:'income',l:'收入'}]"
              :key="opt.v"
              class="filter-chip"
              :class="{ active: filterType === opt.v }"
              @click="filterType = opt.v"
            >{{ opt.l }}</button>
            <span class="filter-sep"></span>
            <button
              v-for="cat in availableFilterCategories"
              :key="'fc-'+cat"
              class="filter-chip"
              :class="{ active: filterCategory === cat }"
              @click="toggleFilterCategory(cat)"
            >{{ cat }}</button>
            <span v-if="availableFilterTags.length > 0" class="filter-sep"></span>
            <button
              v-for="tag in availableFilterTags"
              :key="'ft-'+tag"
              class="filter-chip"
              :class="{ active: filterTag === tag }"
              @click="toggleFilterTag(tag)"
            >#{{ tag }}</button>
          </div>
          <button v-if="hasActiveFilters" class="btn-clear-filters" @click="clearFilters">
            清空筛选
          </button>
        </div>
      </div>

      <div v-if="billGroups.length === 0 && !hasActiveFilters" class="empty-ledger">
        <div v-if="showDeleted">
          <p class="empty-ledger-title">暂无已作废账单</p>
          <p class="empty-ledger-desc">删除后的记录会出现在设置中的回收站</p>
        </div>
        <div v-else>
          <p class="empty-ledger-title">暂无账单</p>
          <p class="empty-ledger-desc">这个月还没有记录，先记一笔吧</p>
          <button type="button" class="btn-create" :disabled="loading !== ''" @click="openCreateBill">
            + 记一笔
          </button>
        </div>
      </div>

      <div v-else-if="billGroups.length === 0 && hasActiveFilters" class="empty-ledger">
        <p class="empty-ledger-title">没有找到匹配账单</p>
        <p class="empty-ledger-desc">可以换个关键词或清空筛选试试</p>
        <button type="button" class="btn-clear-filters" @click="clearFilters">清空筛选</button>
      </div>

      <div v-else class="ledger-list-panel">
      <section v-for="group in billGroups" :key="group.date" class="bill-day">
        <div class="bill-day-header">
          <span class="bill-day-date">{{ group.label }}</span>
          <span class="bill-day-summary">
            <span v-if="group.expense > 0" class="daily-expense">支出 {{ formatMoney(group.expense) }}</span>
            <span v-if="group.expense > 0 && group.income > 0" class="daily-sep">·</span>
            <span v-if="group.income > 0" class="daily-income">收入 {{ formatMoney(group.income) }}</span>
            <span class="daily-count">{{ group.bills.length }} 笔</span>
          </span>
        </div>
        <div class="bill-list">
          <article
            v-for="bill in group.bills"
            :key="bill.id"
            class="bill-card"
            :class="{ selected: selectedBillId === bill.id, deleted: showDeleted }"
            @click="openBillDetail(bill.id)"
          >
            <div class="bill-card-left">
              <div class="bill-category">{{ categoryText(bill) }}</div>
              <div v-if="bill.merchant || bill.note" class="bill-detail">
                <span v-if="bill.merchant" class="bill-merchant">{{ bill.merchant }}</span>
                <span v-if="bill.merchant && bill.note" class="bill-dot">·</span>
                <span v-if="bill.note" class="bill-note">{{ bill.note }}</span>
              </div>
              <div v-if="listTags(bill).length > 0" class="bill-tags">
                <span v-for="tag in listTags(bill).slice(0, 3)" :key="tag" class="tag-chip">{{ tag }}</span>
                <span v-if="listTags(bill).length > 3" class="tag-chip tag-more">+{{ listTags(bill).length - 3 }}</span>
                <span v-if="showDeleted" class="tag-chip tag-deleted">已作废</span>
              </div>
            </div>
            <div class="bill-card-right">
              <span class="bill-amount" :class="amountClass(bill.type)">
                {{ formatBillAmount(bill) }}
              </span>
            </div>
          </article>
        </div>
      </section>
      </div><!-- .ledger-list-panel -->
    </section>

    <section class="details-grid">
      <details class="output-panel">
        <summary>
          <span>verify 输出</span>
          <button type="button" :disabled="loading !== ''" @click.prevent="runVerify">
            {{ loading === 'verify' ? '验证中' : '运行 verify' }}
          </button>
        </summary>
        <pre>{{ verifyOutput || '尚未运行 verify' }}</pre>
      </details>

      <details class="output-panel">
        <summary>
          <span>backup 输出</span>
        </summary>
        <pre>{{ backupOutput || '尚未运行手动同步' }}</pre>
      </details>
    </section>
  </main>
    </div><!-- .main-area -->
  </div><!-- .app-shell -->

  <aside
    v-if="detailPanelStatus !== 'closed' || showSettingsPanel"
    class="detail-panel"
    aria-label="账单详情"
    @keydown="onDetailPanelKeydown"
  >
    <template v-if="showSettingsPanel">
      <!-- Deleted Bills Sub-View -->
      <template v-if="settingsView === 'deleted-bills'">
        <header class="detail-panel-header">
          <div class="detail-header-row">
            <div>
              <p class="eyebrow">Settings · 数据维护</p>
              <h2>已删除账单</h2>
              <p class="detail-header-subtitle">这些账单不会出现在默认账单流中，可以在这里恢复</p>
            </div>
            <button type="button" class="icon-button" @click="goBackToSettings">←</button>
          </div>
        </header>
        <div class="detail-body">
          <!-- Loading -->
          <p v-if="deletedBillsLoading" class="soft-hint">正在加载...</p>

          <!-- Error -->
          <div v-else-if="deletedBillsError" class="settings-error">
            <span>{{ deletedBillsError }}</span>
            <button type="button" class="btn-retry" @click="openDeletedBillsView">重试</button>
          </div>

          <!-- Empty -->
          <div v-else-if="deletedBillsList.length === 0" class="deleted-empty">
            <p class="deleted-empty-title">暂无已删除账单</p>
            <p class="soft-hint">删除后的记录会出现在这里，方便后续恢复</p>
          </div>

          <!-- List -->
          <div v-else class="deleted-list">
            <article
              v-for="bill in deletedBillsList"
              :key="bill.id"
              class="deleted-card"
            >
              <div class="deleted-card-main">
                <div class="deleted-card-header">
                  <span class="deleted-date">{{ bill.bill_time.slice(0, 10) }}</span>
                  <span :class="bill.type === 'income' ? 'income-amount' : ''" class="deleted-amount">
                    {{ formatMoney(bill.amount) }}
                  </span>
                </div>
                <div class="deleted-card-info">
                  <span class="category-chip">{{ categoryText(bill) }}</span>
                  <span v-if="bill.merchant" class="tag-chip">{{ bill.merchant }}</span>
                  <template v-for="(tag, idx) in listTags(bill)" :key="`d-${bill.id}-${tag}`">
                    <span v-if="idx < 2" class="tag-chip">{{ tag }}</span>
                  </template>
                  <span v-if="bill.note" class="deleted-note">{{ bill.note }}</span>
                </div>
              </div>
              <button
                type="button"
                class="btn-restore-small"
                :disabled="restoringBillId === bill.id"
                @click="restoreDeletedBill(bill.id)"
              >
                {{ restoringBillId === bill.id ? '恢复中' : '恢复' }}
              </button>
            </article>
          </div>
        </div>
        <footer class="detail-footer">
          <span class="detail-status">{{ deletedBillsList.length }} 条已删除账单</span>
          <button type="button" @click="goBackToSettings">返回设置</button>
        </footer>
      </template>

      <!-- Main Settings View -->
      <template v-else>
        <header class="detail-panel-header">
          <div class="detail-header-row">
            <div>
              <p class="eyebrow">Configuration</p>
              <h2>SETTINGS</h2>
            </div>
            <button type="button" class="icon-button" @click="closeSettings">x</button>
          </div>
        </header>
        <div class="detail-body">
          <section class="settings-section">
            <h3 class="settings-section-title">配置状态</h3>
            <div v-if="configStatus">
              <div class="settings-status-row">
                <span>总状态</span>
                <span :class="configStatus.overall_status === 'ok' ? 'status-ok' : 'status-err'">
                  {{ configStatus.overall_status === 'ok' ? '正常' : '异常' }}
                </span>
              </div>
              <ul v-if="configStatus.errors.length > 0" class="config-error-list">
                <li v-for="(err, idx) in configStatus.errors" :key="idx">{{ err }}</li>
              </ul>
            </div>
            <p v-else class="soft-hint">正在加载配置...</p>
          </section>
          <section class="settings-section">
            <h3 class="settings-section-title">路径配置</h3>
            <div class="settings-form-fields">
              <label>
                <span>account_book_exe</span>
                <input v-model="settingsForm.account_book_exe" type="text" />
              </label>
              <label>
                <span>db_path</span>
                <input v-model="settingsForm.db_path" type="text" />
              </label>
              <label>
                <span>backup_repo</span>
                <input v-model="settingsForm.backup_repo" type="text" />
              </label>
              <label>
                <span>default_month</span>
                <input v-model="settingsForm.default_month" type="text" placeholder="current 或 YYYY-MM" />
              </label>
            </div>
          </section>
          <section class="settings-section">
            <h3 class="settings-section-title">数据维护</h3>
            <div class="settings-link-list">
              <button type="button" class="settings-link-item" @click="openDeletedBillsView">
                <span class="settings-link-label">已删除账单</span>
                <span class="settings-link-desc">查看并恢复已删除的记录</span>
                <span class="settings-link-arrow">→</span>
              </button>
              <button
                type="button"
                class="settings-link-item"
                :class="{ 'action-active': maintenanceLoading === 'backup' }"
                :disabled="maintenanceLoading !== ''"
                @click="runMaintenanceBackup"
              >
                <span class="settings-link-label">备份数据</span>
                <span class="settings-link-desc">生成本地备份文件</span>
                <span class="settings-link-action">
                  {{ maintenanceLoading === 'backup' ? '备份中...' : '执行' }}
                </span>
              </button>
              <button
                type="button"
                class="settings-link-item"
                :class="{ 'action-active': maintenanceLoading === 'export' }"
                :disabled="maintenanceLoading !== ''"
                @click="runMaintenanceExport"
              >
                <span class="settings-link-label">导出数据</span>
                <span class="settings-link-desc">导出账单数据，便于查看或迁移</span>
                <span class="settings-link-action">
                  {{ maintenanceLoading === 'export' ? '导出中...' : '执行' }}
                </span>
              </button>
              <button
                type="button"
                class="settings-link-item"
                :class="{ 'action-active': maintenanceLoading === 'report' }"
                :disabled="maintenanceLoading !== ''"
                @click="runMaintenanceReport"
              >
                <span class="settings-link-label">生成报告</span>
                <span class="settings-link-desc">生成当前月份报告</span>
                <span class="settings-link-action">
                  {{ maintenanceLoading === 'report' ? '生成中...' : '执行' }}
                </span>
              </button>
              <button type="button" class="settings-link-item" @click="openDataDir">
                <span class="settings-link-label">打开数据目录</span>
                <span class="settings-link-desc">查看 SQLite、配置文件</span>
                <span class="settings-link-arrow">→</span>
              </button>
              <button type="button" class="settings-link-item" @click="openBackupDir">
                <span class="settings-link-label">打开备份目录</span>
                <span class="settings-link-desc">查看本地备份文件</span>
                <span class="settings-link-arrow">→</span>
              </button>
              <button type="button" class="settings-link-item" @click="openExportDir">
                <span class="settings-link-label">打开导出目录</span>
                <span class="settings-link-desc">查看导出数据文件</span>
                <span class="settings-link-arrow">→</span>
              </button>
              <button type="button" class="settings-link-item" @click="openReportDir">
                <span class="settings-link-label">打开报告目录</span>
                <span class="settings-link-desc">查看生成的报告文件</span>
                <span class="settings-link-arrow">→</span>
              </button>
            </div>
          </section>

          <section v-if="maintenanceResult || maintenanceError" class="settings-section">
            <p v-if="maintenanceResult" class="settings-message"><pre class="maintenance-output">{{ maintenanceResult }}</pre></p>
            <p v-if="maintenanceError" class="settings-error">{{ maintenanceError }}</p>
          </section>

          <section class="settings-section">
            <h3 class="settings-section-title">最近操作</h3>
            <div class="recent-ops">
              <div class="recent-op-row">
                <span class="recent-op-label">手动同步</span>
                <span v-if="recentOperations.sync" class="recent-op-time">{{ formatDateTime(recentOperations.sync) }}</span>
                <span v-else class="recent-op-none">暂无记录</span>
              </div>
              <div class="recent-op-row">
                <span class="recent-op-label">本地备份</span>
                <span v-if="recentOperations.backup" class="recent-op-time">{{ formatDateTime(recentOperations.backup) }}</span>
                <span v-else class="recent-op-none">暂无记录</span>
              </div>
              <div class="recent-op-row">
                <span class="recent-op-label">数据导出</span>
                <span v-if="recentOperations.export" class="recent-op-time">{{ formatDateTime(recentOperations.export) }}</span>
                <span v-else class="recent-op-none">暂无记录</span>
              </div>
              <div class="recent-op-row">
                <span class="recent-op-label">报告生成</span>
                <span v-if="recentOperations.report" class="recent-op-time">{{ formatDateTime(recentOperations.report) }}</span>
                <span v-else class="recent-op-none">暂无记录</span>
              </div>
            </div>
          </section>

          <p v-if="settingsError" class="settings-error">{{ settingsError }}</p>
          <p v-if="settingsMessage" class="settings-message">{{ settingsMessage }}</p>
        </div>
        <footer class="detail-footer">
          <span class="detail-status"></span>
          <div class="detail-footer-actions">
            <button type="button" :disabled="loading !== ''" @click="loadConfig">重新加载</button>
            <button type="button" @click="closeSettings">关闭</button>
            <button type="button" :disabled="loading !== ''" @click="saveSettings">
              {{ loading === 'settings' ? '保存中' : '保存配置' }}
            </button>
          </div>
        </footer>
      </template>
    </template>
    <template v-else>
      <header class="detail-panel-header">
      <template v-if="detailPanelStatus === 'loading'">
        <div class="detail-header-row">
          <div>
            <p class="eyebrow">Bill detail</p>
            <h2>正在加载账单详情...</h2>
          </div>
          <button type="button" class="icon-button" @click="closeDetailPanel">x</button>
        </div>
      </template>
      <template v-else-if="detailPanelStatus === 'create'">
        <div class="detail-header-row">
          <div>
            <p class="eyebrow">New bill</p>
            <h2>新增账单</h2>
          </div>
          <button type="button" class="icon-button" @click="closeDetailPanel">x</button>
        </div>
      </template>
      <template v-else-if="billDetail">
        <div class="detail-header-row">
          <div>
            <p class="eyebrow">Bill detail</p>
            <h2>{{ billDetail.display_title }}</h2>
            <p v-if="billDetail.display_subtitle" class="detail-header-subtitle">
              {{ billDetail.display_subtitle }}
            </p>
          </div>
          <button type="button" class="icon-button" @click="closeDetailPanel">x</button>
        </div>
        <div class="detail-amount" :class="amountClass(billDetail.type)">
          {{ formatMoney(billDetail.amount) }}
        </div>
        <p class="detail-subtitle">
          {{ formatDateTime(billDetail.bill_time) }} · {{ typeLabel(billDetail.type) }}
        </p>
      </template>
    </header>

    <div class="detail-body">
      <div class="detail-form">
        <div class="form-group-label">基础信息</div>

        <label>
          <span>备注 / 描述</span>
          <textarea v-model="detailForm.note" rows="3" placeholder="输入备注..." />
        </label>

        <section class="selector-block">
          <div class="field-label">金额</div>
          <input
            ref="amountInputRef"
            v-model.number="detailForm.amount"
            type="number"
            step="0.01"
            min="0"
            placeholder="0.00"
            class="amount-input"
          />
        </section>

        <section class="selector-block">
          <div class="field-label">时间</div>
          <input v-model="detailForm.billTime" type="datetime-local" class="time-input" />
        </section>

        <div class="form-group-label">分类信息</div>

        <section class="selector-block">
          <div class="field-label">类型</div>
          <div class="chip-picker">
            <button
              v-for="bt in billTypes"
              :key="bt.value"
              type="button"
              class="choice-chip"
              :class="{ selected: detailForm.billType === bt.value }"
              @click="detailForm.billType = bt.value"
            >
              {{ bt.label }}
            </button>
          </div>
        </section>

        <section class="selector-block">
          <div class="field-label">商家 / 渠道</div>
          <input v-model="detailForm.merchant" type="text" placeholder="输入商家名称" class="merchant-input" />
          <div class="chip-picker">
            <button
              v-for="m in commonMerchants"
              :key="m"
              type="button"
              class="choice-chip"
              :class="{ selected: detailForm.merchant === m }"
              @click="detailForm.merchant = m"
            >
              {{ m }}
            </button>
          </div>
          <div v-if="createMode && recentMerchants.length > 0" class="recent-chips">
            <span class="recent-chips-label">最近商家</span>
            <button
              v-for="m in recentMerchants"
              :key="'rm-'+m"
              type="button"
              class="choice-chip recent-choice-chip"
              :class="{ selected: detailForm.merchant === m }"
              @click="detailForm.merchant = m"
            >
              {{ m }}
            </button>
          </div>
        </section>

        <section class="selector-block">
          <div class="field-label">分类</div>
          <p v-if="hasMissingCategory" class="soft-hint">
            当前分类不在分类树中，已临时选中可用分类。
          </p>
          <div v-if="createMode && recentCategoryPairs.length > 0" class="recent-chips">
            <span class="recent-chips-label">最近使用</span>
            <button
              v-for="pair in recentCategoryPairs"
              :key="pair.category + '/' + pair.subCategory"
              type="button"
              class="choice-chip recent-choice-chip"
              :class="{ selected: detailForm.category === pair.category && detailForm.subCategory === pair.subCategory }"
              @click="selectRecentCategoryPair(pair)"
            >
              {{ pair.category }} / {{ pair.subCategory }}
            </button>
          </div>
          <div class="chip-picker">
            <button
              v-for="category in orderedCategories"
              :key="category.id"
              type="button"
              class="choice-chip"
              :class="{ selected: detailForm.category === category.name }"
              @click="selectCategory(category.name)"
            >
              {{ category.name }}
            </button>
          </div>
        </section>

        <section class="selector-block">
          <div class="field-label">子分类</div>
          <div class="chip-picker">
            <button
              v-for="subCategory in orderedSubCategories"
              :key="subCategory.id"
              type="button"
              class="choice-chip"
              :class="{ selected: detailForm.subCategory === subCategory.name }"
              @click="selectSubCategory(subCategory.name)"
            >
              {{ subCategory.name }}
            </button>
          </div>
        </section>

        <div class="form-group-label">标签</div>

        <section class="selector-block">
          <div class="chip-preview">
            <span v-for="tag in detailForm.tags" :key="tag" class="tag-chip editable-chip">
              {{ tag }}
              <button type="button" @click="removeTag(tag)">x</button>
            </span>
          </div>
          <input v-model="tagSearch" class="tag-search" type="search" placeholder="搜索标签..." @focus="onTagSearchFocus" @blur="onTagSearchBlur" />
          <div v-if="tagSearchActive || tagSearch" class="tag-candidate-list">
            <section v-for="group in tagCandidateGroups" :key="group.name" class="tag-group">
              <h3>{{ group.label }}</h3>
              <div class="chip-picker">
                <button
                  v-for="tag in group.tags"
                  :key="tag.id"
                  type="button"
                  class="choice-chip"
                  @click="addTag(tag.name)"
                >
                  {{ tag.name }}
                </button>
              </div>
            </section>
            <p v-if="tagCandidateGroups.length === 0" class="soft-hint">没有匹配标签</p>
          </div>
        </section>
      </div>

      <details v-if="billDetail" class="raw-block">
        <summary>原始数据</summary>
        <dl>
          <div>
            <dt>原始分类</dt>
            <dd>{{ billDetail.raw_category || '无' }} / {{ billDetail.raw_sub_category || '无' }}</dd>
          </div>
          <div>
            <dt>原始标签</dt>
            <dd>
              <span v-if="billDetail.raw_tags.length === 0">无</span>
              <span v-for="tag in billDetail.raw_tags" :key="tag" class="tag-chip">{{ tag }}</span>
            </dd>
          </div>
          <div>
            <dt>创建时间</dt>
            <dd>{{ formatDateTime(billDetail.created_at) }}</dd>
          </div>
        </dl>
      </details>

      <section v-if="billDetail && detailPanelStatus === 'edit'" class="detail-more-actions">
        <details class="more-actions-block">
          <summary>更多操作</summary>
          <template v-if="showDeleted">
            <p class="soft-hint">此账单已作废，恢复后将重新参与统计。</p>
            <button type="button" class="btn-restore" :disabled="loading !== ''" @click="restoreBill(billDetail.id)">
              {{ loading === 'restore' ? '恢复中' : '恢复账单' }}
            </button>
          </template>
          <template v-else>
            <template v-if="confirmDeleteId === billDetail.id">
              <p class="soft-hint">确认作废这笔账单吗？作废后不会参与统计，但仍可恢复。</p>
              <div class="confirm-actions">
                <button type="button" @click="cancelDelete">取消</button>
                <button type="button" class="btn-danger" :disabled="loading !== ''" @click="softDeleteBill(billDetail.id)">
                  {{ loading === 'delete' ? '作废中' : '确认作废' }}
                </button>
              </div>
            </template>
            <template v-else>
              <button type="button" class="btn-danger" @click="confirmDelete(billDetail.id)">作废账单</button>
            </template>
          </template>
        </details>
      </section>
    </div>

    <footer class="detail-footer">
      <span class="detail-status" :class="{ dirty: hasUnsavedChanges }">{{ footerStatus }}</span>
      <div v-if="createMode" class="detail-footer-actions">
        <button type="button" @click="closeDetailPanel">取消</button>
        <button
          type="button"
          :disabled="loading !== ''"
          @click="saveBillDetail(true)"
        >
          {{ loading === 'save-detail' ? '保存中' : '保存并关闭' }}
        </button>
        <button
          type="button"
          :disabled="loading !== ''"
          @click="saveBillDetail(false)"
        >
          {{ loading === 'save-detail' ? '保存中' : '保存' }}
        </button>
      </div>
      <div v-else class="detail-footer-actions">
        <button type="button" @click="closeDetailPanel">取消</button>
        <button type="button" :disabled="loading !== '' || !hasUnsavedChanges" @click="saveBillDetail">
          {{ loading === 'save-detail' ? '保存中' : '保存修改' }}
        </button>
      </div>
    </footer>
    </template>
  </aside>
</template>
