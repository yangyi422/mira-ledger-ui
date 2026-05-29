<script setup>
import { computed, nextTick, onErrorCaptured, onMounted, ref, watch } from 'vue'
import AppIcon from './components/AppIcon.vue'
import IconPicker from './components/IconPicker.vue'
import { getCategoryIconKey } from './constants/icons.js'
import {
  GetAllTags,
  GetBillDetail,
  GetBills,
  GetCategoryTree,
  GetConfig,
  GetConfigStatus,
  ListCategories,
  ListTags,
  ListMerchants,
  CreateCategory,
  CreateTag,
  CreateMerchant,
  UpdateCategory,
  UpdateTag,
  UpdateMerchant,
  SetCategoryActive,
  SetTagActive,
  SetMerchantActive,
  DeleteCategory,
  DeleteTag,
  DeleteMerchant,
  MergeTags,
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
  SearchBills,
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
  project_hint: '项目线索',
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
const searchedBills = ref([])
const searchLoading = ref(false)
const searchPerformed = ref(false)

let searchTimer = null
function onSearchInput() {
  clearTimeout(searchTimer)
  const q = searchQuery.value.trim()
  if (!q) {
    searchedBills.value = []
    searchPerformed.value = false
    return
  }
  searchTimer = setTimeout(() => performSearch(q), 300)
}

async function performSearch(q) {
  if (!q) return
  searchLoading.value = true
  searchPerformed.value = true
  try {
    searchedBills.value = (await SearchBills(q)) || []
  } catch (e) {
    errorMessage.value = e?.message || String(e)
  } finally {
    searchLoading.value = false
  }
}

function refreshAfterMutation() {
  if (isSearching.value) {
    performSearch(searchQuery.value.trim())
  }
}

const filterType = ref('all')   // 'all' | 'expense' | 'income'
const filterCategory = ref('')  // category name or ''
const filterTag = ref('')       // tag name or ''

function onTagSearchFocus() { tagSearchActive.value = true }
function onTagSearchBlur() { setTimeout(() => { tagSearchActive.value = false }, 200) }
const showDeleted = ref(false)
const showMoreFilters = ref(false)
const confirmDeleteId = ref(null)
const deletedBills = ref([])
const createMode = ref(false)
const detailPanelStatus = ref('ledger') // 'ledger' | 'create' | 'loading' | 'edit'
const errorMessage = ref('')
const detailMessage = ref('')
const insightView = ref(null)  // null | 'categories' | 'merchants' | 'expenses'
const loading = ref('')
const lastCreatedBill = ref(null)  // { billType, category, subCategory, billTime } from last create
const amountInputRef = ref(null)

const currentView = ref('dashboard') // 'dashboard' | 'categories' | 'tags' | 'merchants'
const appConfig = ref(null)
const configError = ref('')
const configStatus = ref(null)

const showSettings = ref(false)
const showSettingsPanel = ref(false)
const settingsView = ref('main') // 'main' | 'deleted-bills' | 'category-manager' | 'tag-manager'
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

const managedCategories = ref([])
const categoryManageType = ref('expense')
const categoryManageSearch = ref('')
const categoryEditorOpen = ref(false)
const categoryManagerLoading = ref(false)
const categoryManagerError = ref('')
const categoryManagerMessage = ref('')
const categoryEditingId = ref(null)
const categoryForm = ref({ parent_id: 0, name: '', type: 'expense', sort_order: 0, icon_key: '' })
const inlineEditingId = ref(null)
const inlineDraft = ref({ name: '', icon_key: '' })
const inlineIconPickerOpen = ref(false)

const managedTags = ref([])
const tagManageGroup = ref('all')
const tagManageSearch = ref('')
const tagShowInactiveOnly = ref(false)
const tagShowUngroupedOnly = ref(false)
const tagEditorOpen = ref(false)
const tagManagerLoading = ref(false)
const tagManagerError = ref('')
const tagManagerMessage = ref('')
const tagEditingId = ref(null)
const tagForm = ref({ name: '', group_name: 'content', sort_order: 0 })
const mergeSourceTagId = ref(null)
const mergeTargetTagId = ref('')

const managedMerchants = ref([])
const merchantManageSearch = ref('')
const merchantEditorOpen = ref(false)
const merchantManagerLoading = ref(false)
const merchantManagerError = ref('')
const merchantManagerMessage = ref('')
const merchantEditingId = ref(null)
const merchantForm = ref({ name: '', sort_order: 0 })

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

function navigateTo(view) {
  currentView.value = view
  closeDetailPanel()
  closeInsightView()
  if (showSettingsPanel.value) {
    closeSettings()
  }
  if (view === 'categories') {
    categoryManageSearch.value = ''
    categoryEditorOpen.value = false
    resetCategoryForm()
    refreshCategoryManagement()
  } else if (view === 'tags') {
    tagManageGroup.value = 'all'
    tagManageSearch.value = ''
    tagShowInactiveOnly.value = false
    tagShowUngroupedOnly.value = false
    tagEditorOpen.value = false
    resetTagForm()
    cancelMergeTag()
    refreshTagManagement()
  } else if (view === 'merchants') {
    merchantManageSearch.value = ''
    merchantEditorOpen.value = false
    resetMerchantForm()
    refreshMerchantManagement()
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

async function refreshReferenceDataAfterManagement() {
  await Promise.all([loadReferenceData(), refreshDashboard()])
}

async function openCategoryManager() {
  settingsError.value = ''
  settingsMessage.value = ''
  showSettings.value = true
  showSettingsPanel.value = true
  settingsView.value = 'category-manager'
  resetCategoryForm()
  await refreshCategoryManagement()
}

async function refreshCategoryManagement() {
  categoryManagerError.value = ''
  categoryManagerLoading.value = true
  try {
    managedCategories.value = (await ListCategories(true)) || []
  } catch (error) {
    categoryManagerError.value = error?.message || String(error)
  } finally {
    categoryManagerLoading.value = false
  }
}

function resetCategoryForm(parent = null) {
  categoryEditingId.value = null
  categoryForm.value = {
    parent_id: parent?.id || 0,
    name: '',
    type: parent?.type || categoryManageType.value,
    sort_order: 0,
    icon_key: '',
  }
  categoryManagerMessage.value = ''
  categoryManagerError.value = ''
  categoryEditorOpen.value = parent ? true : !categoryEditorOpen.value
}

function toggleCategoryEditor() {
  cancelInlineEdit()
  categoryEditorOpen.value = !categoryEditorOpen.value
  if (!categoryEditorOpen.value) {
    resetCategoryForm()
    categoryEditorOpen.value = false
  }
}

// ── Inline editing ──
function startInlineEdit(category) {
  cancelInlineEdit()
  categoryEditorOpen.value = false
  inlineEditingId.value = category.id
  inlineDraft.value = {
    name: category.name || '',
    icon_key: category.icon_key || getCategoryIconKey(category),
  }
  inlineIconPickerOpen.value = false
}

function cancelInlineEdit() {
  inlineEditingId.value = null
  inlineDraft.value = { name: '', icon_key: '' }
  inlineIconPickerOpen.value = false
}

async function saveInlineEdit(category) {
  const name = inlineDraft.value.name.trim()
  if (!name) {
    categoryManagerError.value = '请输入分类名称'
    return
  }
  categoryManagerLoading.value = true
  categoryManagerError.value = ''
  categoryManagerMessage.value = ''
  try {
    await UpdateCategory({
      id: category.id,
      name,
      sort_order: category.sort_order || 0,
      icon_key: (inlineDraft.value.icon_key || '').trim(),
    })
    categoryManagerMessage.value = '分类已更新'
    cancelInlineEdit()
    await refreshCategoryManagement()
    await refreshReferenceDataAfterManagement()
  } catch (err) {
    categoryManagerError.value = err?.message || String(err)
  } finally {
    categoryManagerLoading.value = false
  }
}

// ── Drag-and-drop for category sorting ──
const dragCategoryId = ref(null)
const dragOverCategoryId = ref(null)
const dragIsChild = ref(false)

function onCategoryDragStart(e, category, isChild) {
  e.stopPropagation()
  dragCategoryId.value = category.id
  dragIsChild.value = isChild
  e.dataTransfer.effectAllowed = 'move'
  e.dataTransfer.setData('text/plain', String(category.id))
}

function onCategoryDragOver(e, category, isChild) {
  e.preventDefault()
  if (!dragCategoryId.value) return
  if (dragIsChild.value !== isChild) return
  dragOverCategoryId.value = category.id
}

function onCategoryDragLeave(e) {
  // Only fire when leaving the bound element (not inner children)
  if (e.currentTarget.contains(e.relatedTarget)) return
  dragOverCategoryId.value = null
}

async function onCategoryDrop(e, targetCategory, isChild) {
  e.stopPropagation()
  dragOverCategoryId.value = null
  if (!dragCategoryId.value) return
  if (dragCategoryId.value === targetCategory.id) {
    dragCategoryId.value = null
    return
  }
  if (dragIsChild.value !== isChild) return

  let list
  if (isChild) {
    // Find parent and reorder children
    const parent = managedCategories.value.find((c) =>
      c.children && c.children.some((ch) => ch.id === dragCategoryId.value)
    )
    if (!parent) { dragCategoryId.value = null; return }
    if (!parent.children.some((ch) => ch.id === targetCategory.id)) { dragCategoryId.value = null; return }
    list = parent.children
  } else {
    list = managedCategoryRoots.value
  }

  const fromIdx = list.findIndex((c) => c.id === dragCategoryId.value)
  const toIdx = list.findIndex((c) => c.id === targetCategory.id)
  if (fromIdx === -1 || toIdx === -1) { dragCategoryId.value = null; return }

  // Reorder in place
  const [moved] = list.splice(fromIdx, 1)
  list.splice(toIdx, 0, moved)

  // Renumber sort_order and persist
  const updates = list.map((c, i) => ({ id: c.id, sort_order: i + 1, name: c.name, icon_key: c.icon_key || '' }))
  categoryManagerLoading.value = true
  try {
    for (const u of updates) {
      await UpdateCategory({
        id: u.id,
        name: u.name,
        sort_order: u.sort_order,
        icon_key: u.icon_key,
      })
    }
    // Refresh local state from server
    await refreshCategoryManagement()
    categoryManagerMessage.value = '排序已更新'
  } catch (err) {
    categoryManagerError.value = err?.message || String(err)
    await refreshCategoryManagement()
  } finally {
    categoryManagerLoading.value = false
    dragCategoryId.value = null
    dragOverCategoryId.value = null
  }
}

function onDragEnd() {
  dragCategoryId.value = null
  dragOverCategoryId.value = null
}

function editCategory(category) {
  categoryEditingId.value = category.id
  categoryForm.value = {
    parent_id: category.parent_id || 0,
    name: category.name || '',
    type: category.type || categoryManageType.value,
    sort_order: category.sort_order || 0,
    icon_key: category.icon_key || '',
  }
  categoryManagerMessage.value = ''
  categoryManagerError.value = ''
}

async function saveManagedCategory() {
  const name = categoryForm.value.name.trim()
  if (!name) {
    categoryManagerError.value = '请输入分类名称'
    return
  }
  categoryManagerLoading.value = true
  categoryManagerError.value = ''
  categoryManagerMessage.value = ''
  try {
    if (categoryEditingId.value) {
      await UpdateCategory({
        id: categoryEditingId.value,
        name,
        sort_order: Number(categoryForm.value.sort_order) || 0,
        icon_key: (categoryForm.value.icon_key || '').trim(),
      })
      categoryManagerMessage.value = '分类已更新'
    } else {
      // Auto-compute sort_order: max sibling sort_order + 1
      const siblings = managedCategories.value.filter(
        (c) => c.type === (categoryForm.value.type || categoryManageType.value) && c.parent_id === (Number(categoryForm.value.parent_id) || 0)
      )
      const maxSort = siblings.reduce((max, c) => Math.max(max, c.sort_order || 0), 0)
      await CreateCategory({
        parent_id: Number(categoryForm.value.parent_id) || 0,
        name,
        type: categoryForm.value.type || categoryManageType.value,
        sort_order: maxSort + 1,
        icon_key: (categoryForm.value.icon_key || '').trim(),
      })
      categoryManagerMessage.value = '分类已创建'
    }
    resetCategoryForm()
    await refreshCategoryManagement()
    await refreshReferenceDataAfterManagement()
  } catch (error) {
    categoryManagerError.value = error?.message || String(error)
  } finally {
    categoryManagerLoading.value = false
  }
}

async function toggleCategoryActive(category) {
  categoryManagerLoading.value = true
  categoryManagerError.value = ''
  categoryManagerMessage.value = ''
  try {
    await SetCategoryActive(category.id, !category.is_active)
    categoryManagerMessage.value = category.is_active ? '分类已停用' : '分类已启用'
    await refreshCategoryManagement()
    await refreshReferenceDataAfterManagement()
  } catch (error) {
    categoryManagerError.value = error?.message || String(error)
  } finally {
    categoryManagerLoading.value = false
  }
}

async function deleteManagedCategory(category) {
  if (category.bill_count > 0 || (category.children && category.children.length > 0)) {
    categoryManagerError.value = '该分类已有账单或子分类，请停用而不是删除'
    return
  }
  categoryManagerLoading.value = true
  categoryManagerError.value = ''
  categoryManagerMessage.value = ''
  try {
    await DeleteCategory(category.id)
    categoryManagerMessage.value = '分类已删除'
    await refreshCategoryManagement()
    await refreshReferenceDataAfterManagement()
  } catch (error) {
    categoryManagerError.value = error?.message || String(error)
  } finally {
    categoryManagerLoading.value = false
  }
}

async function openTagManager() {
  settingsError.value = ''
  settingsMessage.value = ''
  showSettings.value = true
  showSettingsPanel.value = true
  settingsView.value = 'tag-manager'
  resetTagForm()
  await refreshTagManagement()
}

async function refreshTagManagement() {
  tagManagerError.value = ''
  tagManagerLoading.value = true
  try {
    managedTags.value = (await ListTags(true)) || []
  } catch (error) {
    tagManagerError.value = error?.message || String(error)
  } finally {
    tagManagerLoading.value = false
  }
}

function resetTagForm() {
  tagEditingId.value = null
  tagForm.value = { name: '', group_name: 'content', sort_order: 0 }
  tagManagerMessage.value = ''
  tagManagerError.value = ''
}

function toggleTagEditor() {
  tagEditorOpen.value = !tagEditorOpen.value
  if (!tagEditorOpen.value) {
    resetTagForm()
    cancelMergeTag()
  }
}

function openTagEditorForCreate() {
  resetTagForm()
  cancelMergeTag()
  tagEditorOpen.value = true
}

function openTagEditorForEdit(tag) {
  editManagedTag(tag)
  cancelMergeTag()
  tagEditorOpen.value = true
}

function editManagedTag(tag) {
  tagEditingId.value = tag.id
  tagForm.value = {
    name: tag.name || '',
    group_name: tag.group_name || 'content',
    sort_order: tag.sort_order || 0,
  }
  tagManagerMessage.value = ''
  tagManagerError.value = ''
}

async function saveManagedTag() {
  const name = tagForm.value.name.trim()
  const groupName = tagForm.value.group_name.trim() || 'content'
  if (!name) {
    tagManagerError.value = '请输入标签名称'
    return
  }
  tagManagerLoading.value = true
  tagManagerError.value = ''
  tagManagerMessage.value = ''
  try {
    if (tagEditingId.value) {
      await UpdateTag({
        id: tagEditingId.value,
        name,
        group_name: groupName,
        sort_order: Number(tagForm.value.sort_order) || 0,
      })
      tagManagerMessage.value = '标签已更新'
    } else {
      await CreateTag({
        name,
        group_name: groupName,
        sort_order: Number(tagForm.value.sort_order) || 0,
      })
      tagManagerMessage.value = '标签已创建'
    }
    resetTagForm()
    await refreshTagManagement()
    await refreshReferenceDataAfterManagement()
  } catch (error) {
    tagManagerError.value = error?.message || String(error)
  } finally {
    tagManagerLoading.value = false
  }
}

async function toggleManagedTagActive(tag) {
  tagManagerLoading.value = true
  tagManagerError.value = ''
  tagManagerMessage.value = ''
  try {
    await SetTagActive(tag.id, !tag.is_active)
    tagManagerMessage.value = tag.is_active ? '标签已停用' : '标签已启用'
    await refreshTagManagement()
    await refreshReferenceDataAfterManagement()
  } catch (error) {
    tagManagerError.value = error?.message || String(error)
  } finally {
    tagManagerLoading.value = false
  }
}

async function deleteManagedTag(tag) {
  if (tag.use_count > 0) {
    tagManagerError.value = '该标签已有账单使用，请停用或合并'
    return
  }
  tagManagerLoading.value = true
  tagManagerError.value = ''
  tagManagerMessage.value = ''
  try {
    await DeleteTag(tag.id)
    tagManagerMessage.value = '标签已删除'
    await refreshTagManagement()
    await refreshReferenceDataAfterManagement()
  } catch (error) {
    tagManagerError.value = error?.message || String(error)
  } finally {
    tagManagerLoading.value = false
  }
}

function beginMergeTag(tag) {
  mergeSourceTagId.value = tag.id
  mergeTargetTagId.value = ''
  tagManagerMessage.value = ''
  tagManagerError.value = ''
}

function cancelMergeTag() {
  mergeSourceTagId.value = null
  mergeTargetTagId.value = ''
}

async function mergeManagedTags() {
  const targetID = Number(mergeTargetTagId.value)
  if (!mergeSourceTagId.value || !targetID) {
    tagManagerError.value = '请选择要合并到的目标标签'
    return
  }
  tagManagerLoading.value = true
  tagManagerError.value = ''
  tagManagerMessage.value = ''
  try {
    await MergeTags({ source_id: mergeSourceTagId.value, target_id: targetID })
    tagManagerMessage.value = '标签已合并'
    cancelMergeTag()
    await refreshTagManagement()
    await refreshReferenceDataAfterManagement()
  } catch (error) {
    tagManagerError.value = error?.message || String(error)
  } finally {
    tagManagerLoading.value = false
  }
}

async function refreshMerchantManagement() {
  merchantManagerError.value = ''
  merchantManagerLoading.value = true
  try {
    managedMerchants.value = (await ListMerchants(true)) || []
  } catch (error) {
    merchantManagerError.value = error?.message || String(error)
  } finally {
    merchantManagerLoading.value = false
  }
}

function resetMerchantForm() {
  merchantEditingId.value = null
  merchantForm.value = { name: '', sort_order: 0 }
  merchantManagerMessage.value = ''
  merchantManagerError.value = ''
}

function toggleMerchantEditor() {
  merchantEditorOpen.value = !merchantEditorOpen.value
  if (!merchantEditorOpen.value) {
    resetMerchantForm()
  }
}

function editMerchant(merchant) {
  merchantEditingId.value = merchant.id
  merchantForm.value = {
    name: merchant.name || '',
    sort_order: merchant.sort_order || 0,
  }
  merchantManagerMessage.value = ''
  merchantManagerError.value = ''
  merchantEditorOpen.value = true
}

async function saveManagedMerchant() {
  const name = merchantForm.value.name.trim()
  if (!name) {
    merchantManagerError.value = '请输入商家名称'
    return
  }
  merchantManagerLoading.value = true
  merchantManagerError.value = ''
  merchantManagerMessage.value = ''
  try {
    if (merchantEditingId.value) {
      await UpdateMerchant({
        id: merchantEditingId.value,
        name,
        sort_order: Number(merchantForm.value.sort_order) || 0,
      })
      merchantManagerMessage.value = '商家已更新'
    } else {
      await CreateMerchant({
        name,
        sort_order: Number(merchantForm.value.sort_order) || 0,
      })
      merchantManagerMessage.value = '商家已创建'
    }
    resetMerchantForm()
    merchantEditorOpen.value = false
    await refreshMerchantManagement()
    await refreshReferenceDataAfterManagement()
  } catch (error) {
    merchantManagerError.value = error?.message || String(error)
  } finally {
    merchantManagerLoading.value = false
  }
}

async function toggleManagedMerchantActive(merchant) {
  merchantManagerLoading.value = true
  merchantManagerError.value = ''
  merchantManagerMessage.value = ''
  try {
    await SetMerchantActive(merchant.id, !merchant.is_active)
    merchantManagerMessage.value = merchant.is_active ? '商家已停用' : '商家已启用'
    await refreshMerchantManagement()
    await refreshReferenceDataAfterManagement()
  } catch (error) {
    merchantManagerError.value = error?.message || String(error)
  } finally {
    merchantManagerLoading.value = false
  }
}

async function deleteManagedMerchant(merchant) {
  if (merchant.use_count > 0) {
    merchantManagerError.value = '该商家已有账单使用，请停用而不是删除'
    return
  }
  merchantManagerLoading.value = true
  merchantManagerError.value = ''
  merchantManagerMessage.value = ''
  try {
    await DeleteMerchant(merchant.id)
    merchantManagerMessage.value = '商家已删除'
    await refreshMerchantManagement()
    await refreshReferenceDataAfterManagement()
  } catch (error) {
    merchantManagerError.value = error?.message || String(error)
  } finally {
    merchantManagerLoading.value = false
  }
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

function openInsightView(viewType) { insightView.value = viewType }
function closeInsightView() { insightView.value = null }

function openDataDir() { handleOpenPath(getMaintenanceDir('data')) }
function openBackupDir() { handleOpenPath(getMaintenanceDir('backups')) }
function openExportDir() { handleOpenPath(getMaintenanceDir('exports')) }
function openReportDir() { handleOpenPath(getMaintenanceDir('reports')) }

function onDetailPanelKeydown(event) {
  if (showSettingsPanel.value) return
  if (insightView.value) {
    if (event.key === 'Escape') { event.preventDefault(); closeInsightView(); return }
    return
  }

  if (event.key === 'Escape') {
    if (detailPanelStatus.value !== 'ledger') {
      event.preventDefault()
      closeDetailPanel()
      return
    }
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
  { label: '收入', value: formatMoney(dashboardStats.value?.income), trend: null },
  { label: '支出', value: formatMoney(dashboardStats.value?.expense), trend: null },
  { label: '结余', value: formatMoney(dashboardStats.value?.balance), trend: null },
])

const managedCategoryRoots = computed(() =>
  managedCategories.value.filter((category) => category.type === categoryManageType.value),
)

const filteredManagedCategoryRoots = computed(() => {
  const query = categoryManageSearch.value.trim().toLowerCase()
  if (!query) return managedCategoryRoots.value
  return managedCategoryRoots.value.filter((cat) => {
    const matchSelf = cat.name.toLowerCase().includes(query)
    const matchChild = cat.children && cat.children.some((c) => c.name.toLowerCase().includes(query))
    return matchSelf || matchChild
  })
})

const categoryParentOptions = computed(() =>
  managedCategoryRoots.value.map((category) => ({ id: category.id, name: category.name })),
)

const tagGroupOptions = computed(() => {
  const groups = new Set(Object.keys(tagGroupLabels))
  for (const tag of managedTags.value) {
    if (tag.group_name) groups.add(tag.group_name)
  }
  return [...groups].sort((a, b) => groupLabel(a).localeCompare(groupLabel(b), 'zh-CN'))
})

const visibleManagedTags = computed(() => {
  let result = managedTags.value

  if (tagManageGroup.value !== 'all') {
    result = result.filter((tag) => tag.group_name === tagManageGroup.value)
  }

  if (tagShowInactiveOnly.value) {
    result = result.filter((tag) => !tag.is_active)
  }

  if (tagShowUngroupedOnly.value) {
    result = result.filter((tag) => !tag.group_name || tag.group_name === 'content')
  }

  const query = tagManageSearch.value.trim().toLowerCase()
  if (query) {
    result = result.filter((tag) => tag.name.toLowerCase().includes(query))
  }

  return result
})

const groupedVisibleTags = computed(() => {
  const groups = new Map()
  for (const tag of visibleManagedTags.value) {
    const g = tag.group_name || 'content'
    if (!groups.has(g)) groups.set(g, [])
    groups.get(g).push(tag)
  }
  return [...groups.entries()].sort((a, b) => groupLabel(a[0]).localeCompare(groupLabel(b[0]), 'zh-CN'))
})

const filteredManagedMerchants = computed(() => {
  const query = merchantManageSearch.value.trim().toLowerCase()
  if (!query) return managedMerchants.value
  return managedMerchants.value.filter((m) => m.name.toLowerCase().includes(query))
})

const mergeSourceTag = computed(() =>
  managedTags.value.find((tag) => tag.id === mergeSourceTagId.value) || null,
)

const mergeTargetOptions = computed(() =>
  managedTags.value.filter((tag) => tag.id !== mergeSourceTagId.value),
)

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

const allCategories = computed(() => {
  const map = new Map()
  const totalExpense = expenseBills.value.reduce((s, b) => s + b.amount, 0)
  for (const bill of expenseBills.value) {
    const cat = bill.category || '未分类'
    const prev = map.get(cat) || { category: cat, total: 0, count: 0 }
    prev.total += bill.amount
    prev.count++
    map.set(cat, prev)
  }
  return [...map.values()]
    .sort((a, b) => b.total - a.total)
    .map(item => ({ ...item, pct: totalExpense > 0 ? (item.total / totalExpense * 100) : 0 }))
})

const allMerchants = computed(() => {
  const map = new Map()
  for (const bill of expenseBills.value) {
    const m = bill.merchant
    if (!m) continue
    const prev = map.get(m) || { merchant: m, total: 0, count: 0 }
    prev.total += bill.amount
    prev.count++
    map.set(m, prev)
  }
  return [...map.values()].sort((a, b) => b.total - a.total)
})

const allTopExpenses = computed(() => {
  return [...expenseBills.value].sort((a, b) => b.amount - a.amount)
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

const incomeBills = computed(() => bills.value.filter(b => b.type === 'income'))

const ledgerSideTags = computed(() => {
  const map = new Map()
  const source = showDeleted.value ? deletedBills.value : bills.value
  for (const bill of source) {
    for (const tag of (Array.isArray(bill.tags) ? bill.tags : [])) {
      const prev = map.get(tag) || { name: tag, count: 0 }
      prev.count++
      map.set(tag, prev)
    }
  }
  return [...map.values()].sort((a, b) => b.count - a.count).slice(0, 6)
})

const ledgerSideMerchants = computed(() => {
  const map = new Map()
  const source = showDeleted.value ? deletedBills.value : bills.value
  for (const bill of source) {
    const m = bill.merchant
    if (!m) continue
    const prev = map.get(m) || { name: m, count: 0 }
    prev.count++
    map.set(m, prev)
  }
  return [...map.values()].sort((a, b) => b.count - a.count).slice(0, 5)
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

const recentTags = computed(() => {
  const seen = new Set()
  const tags = []
  for (const bill of bills.value) {
    for (const tag of (Array.isArray(bill.tags) ? bill.tags : [])) {
      if (seen.has(tag) || selectedTagSet.value.has(tag)) continue
      seen.add(tag)
      tags.push(tag)
      if (tags.length >= 8) return tags
    }
  }
  return tags
})

const allAvailableTags = computed(() => {
  const recentSet = new Set(recentTags.value)
  return allTags.value
    .filter(t => !selectedTagSet.value.has(t.name) && !recentSet.has(t.name))
    .map(t => t.name)
    .slice(0, 20)
})

const tagSearchText = computed(() => tagSearch.value.trim())
const tagSearchQuery = computed(() => tagSearchText.value.toLowerCase())

const filteredRecentTags = computed(() => {
  if (!tagSearchQuery.value) return recentTags.value
  return recentTags.value.filter(t => t.toLowerCase().includes(tagSearchQuery.value))
})

const filteredAvailableTags = computed(() => {
  if (!tagSearchQuery.value) return allAvailableTags.value
  return allAvailableTags.value.filter(t => t.toLowerCase().includes(tagSearchQuery.value))
})

const showCreateTag = computed(() => {
  if (!tagSearchText.value) return false
  return !detailForm.value.tags.some((tag) => tag.toLowerCase() === tagSearchQuery.value)
    && !allTags.value.some((tag) => tag.name.toLowerCase() === tagSearchQuery.value)
})

async function createTagFromSearch() {
  const name = tagSearchText.value
  if (!name) return
  const existing = allTags.value.find((tag) => tag.name.toLowerCase() === name.toLowerCase())
  const tagName = existing?.name || name
  if (!selectedTagSet.value.has(tagName)) {
    if (!existing) {
      try {
        await CreateTag({ name: tagName, group_name: 'content', sort_order: 0 })
        await loadReferenceData()
      } catch (error) {
        detailMessage.value = error?.message || String(error)
        return
      }
    }
    detailForm.value.tags = [...detailForm.value.tags, tagName]
  }
  tagSearch.value = ''
}

function toggleTagFromList(name) {
  if (selectedTagSet.value.has(name)) {
    removeTag(name)
  } else {
    addTag(name)
  }
}

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

const isSearching = computed(() => searchQuery.value.trim().length > 0)
const displayBills = computed(() => {
  if (showDeleted.value) return deletedBills.value
  if (isSearching.value) return searchedBills.value
  return filteredBills.value
})

const availableFilterCategories = computed(() => {
  const cats = new Map()
  for (const bill of bills.value) {
    const c = bill.category || '未分类'
    if (!cats.has(c)) cats.set(c, true)
  }
  return [...cats.keys()]
})

const visibleFilterCategories = computed(() => {
  return showMoreFilters.value
    ? availableFilterCategories.value
    : availableFilterCategories.value.slice(0, 5)
})

const hasMoreFilterCategories = computed(() => {
  return availableFilterCategories.value.length > 5
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
  showDeleted.value = false
  deletedBills.value = []
  searchQuery.value = ''
  filterType.value = 'all'
  filterCategory.value = ''
  filterTag.value = ''
  lastCreatedBill.value = null
  closeInsightView()
  detailPanelStatus.value = 'ledger'
})

watch([filteredBills, showDeleted], () => {
  if (selectedBillId.value && !displayBills.value.some(b => b.id === selectedBillId.value)) {
    closeDetailPanel()
  }
})

watch(() => detailForm.value.billType, () => {
  normalizeCategorySelection()
})

watch(categoryManageType, () => {
  if (!categoryEditingId.value) {
    categoryForm.value = { ...categoryForm.value, parent_id: 0, type: categoryManageType.value }
  }
})

const footerStatus = computed(() => {
  if (detailPanelStatus.value === 'ledger') {
    return ''
  }
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
  refreshAfterMutation()
}

async function openBillDetail(id) {
  console.log('[openBillDetail] bill.id =', id, 'showDeleted =', showDeleted.value, 'createMode =', createMode.value)
  insightView.value = null
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
    detailPanelStatus.value = 'ledger'
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

function openCreateBill(dateKey = '') {
  createMode.value = true
  detailPanelStatus.value = 'create'
  selectedBillId.value = null
  billDetail.value = null
  originalDetail.value = null
  detailMessage.value = ''
  tagSearch.value = ''

  let billTime
  if (dateKey) {
    const now = new Date()
    const pad = (n) => String(n).padStart(2, '0')
    billTime = dateKey + 'T' + pad(now.getHours()) + ':' + pad(now.getMinutes())
  } else {
    billTime = lastCreatedBill.value?.billTime || nowLocalDateTime()
  }

  detailForm.value = {
    amount: null,
    billTime: billTime,
    billType: lastCreatedBill.value?.billType || 'expense',
    category: lastCreatedBill.value?.category || '',
    subCategory: lastCreatedBill.value?.subCategory || '',
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
  detailPanelStatus.value = 'ledger'
  selectedBillId.value = null
  billDetail.value = null
  originalDetail.value = null
  confirmDeleteId.value = null
  detailMessage.value = ''
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
  if (detailPanelStatus.value !== 'create' && detailPanelStatus.value !== 'edit') {
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

function formatShortDateTime(value) {
  if (!value) return '暂无'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '暂无'
  const pad = (n) => String(n).padStart(2, '0')
  return `${pad(date.getMonth()+1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
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
        <button class="sidebar-item" :class="{ active: currentView === 'dashboard' }" @click="navigateTo('dashboard')">
          <span class="sidebar-icon"><AppIcon name="overview" :size="16" /></span> 概览
        </button>
        <button class="sidebar-item">
          <span class="sidebar-icon"><AppIcon name="ledger" :size="16" /></span> 账单流
        </button>
        <button class="sidebar-item" @click="openSettings">
          <span class="sidebar-icon"><AppIcon name="settings" :size="16" /></span> 设置
        </button>
        <div class="sidebar-nav-sep"></div>
        <span class="sidebar-section-label">功能</span>
        <button class="sidebar-item">
          <span class="sidebar-icon"><AppIcon name="calendar" :size="16" /></span> 日历视图
        </button>
        <button class="sidebar-item">
          <span class="sidebar-icon"><AppIcon name="reports" :size="16" /></span> 报表
        </button>
        <button class="sidebar-item" :class="{ active: currentView === 'categories' }" @click="navigateTo('categories')">
          <span class="sidebar-icon"><AppIcon name="category" :size="16" /></span> 分类
        </button>
        <button class="sidebar-item" :class="{ active: currentView === 'merchants' }" @click="navigateTo('merchants')">
          <span class="sidebar-icon"><AppIcon name="merchant" :size="16" /></span> 商家
        </button>
        <button class="sidebar-item" :class="{ active: currentView === 'tags' }" @click="navigateTo('tags')">
          <span class="sidebar-icon"><AppIcon name="tag" :size="16" /></span> 标签
        </button>
      </nav>
      <div class="sidebar-footer">
        <span class="sidebar-section-label">状态</span>
        <div class="sidebar-status-card">
          <p class="sidebar-status-title">数据已同步</p>
          <div class="sidebar-status-row">
            <span class="sidebar-status-label">账单</span>
            <span class="sidebar-status-value">有效 {{ dashboardStats?.active_bills ?? '--' }} &middot; 作废 {{ dashboardStats?.deleted_bills ?? '--' }}</span>
          </div>
          <div class="sidebar-status-row">
            <span class="sidebar-status-label">配置</span>
            <span v-if="configStatus" class="sidebar-status-value" :class="{ 'sidebar-status-err': configStatus.overall_status === 'error' }">
              {{ configStatus.overall_status === 'error' ? '异常' : '正常' }}
            </span>
            <span v-else class="sidebar-status-value">--</span>
          </div>
        </div>
      </div>
    </aside>

    <!-- Main Workspace -->
    <div class="main-workspace">
    <main class="dashboard" v-if="currentView === 'dashboard'">
    <header class="topbar">
      <div class="topbar-left">
        <p class="eyebrow">Overview</p>
        <h1>概览</h1>
        <p class="topbar-subtitle">掌控你的个人财务</p>
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
        <button type="button" class="btn-toggle-deleted" :class="{ active: showDeleted }" @click="toggleShowDeleted">
          已作废
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
        <div class="kpi-title">{{ card.label }}</div>
        <div class="kpi-value">{{ card.value }}</div>
        <div v-if="card.trend" class="kpi-trend">{{ card.trend }}</div>
      </article>
    </section>

    <div class="section-divider"><span class="section-divider-label">统计</span></div>

    <section class="stats-grid" aria-label="月度统计">
      <!-- 本月花在哪 -->
      <article class="stats-module">
        <div class="stats-module-header">
          <h3 class="stats-module-title">本月花在哪</h3>
          <button class="stats-view-all" @click="openInsightView('categories')">查看全部 &rarr;</button>
        </div>
        <div v-if="categoryTop5.length > 0" class="stats-list">
          <div v-for="item in categoryTop5" :key="item.category" class="category-row">
            <span class="category-name">{{ item.category }}</span>
            <span class="category-bar"><span class="category-bar-fill" :style="{ width: Math.round(item.pct) + '%' }"></span></span>
            <span class="category-amount">{{ formatMoney(item.total) }}</span>
            <span class="category-percent">{{ Math.round(item.pct) }}%</span>
          </div>
        </div>
        <p v-else class="stats-empty">暂无分类支出</p>
      </article>

      <!-- 本月概况 -->
      <article class="stats-module">
        <div class="stats-module-header">
          <h3 class="stats-module-title">本月概况</h3>
        </div>
        <div class="summary-list">
          <div class="summary-row">
            <span class="summary-label">日均支出</span>
            <span class="summary-value">{{ formatMoney(dailyStats.dailyAvg) }}</span>
          </div>
          <div class="summary-row">
            <span class="summary-label">总支出</span>
            <span class="summary-value">{{ formatMoney(dailyStats.totalExpense) }}</span>
          </div>
          <div class="summary-row">
            <span class="summary-label">支出笔数</span>
            <span class="summary-value">{{ expenseBills.length }} 笔</span>
          </div>
          <div v-if="dailyStats.isCurrentMonth" class="summary-row">
            <span class="summary-label">已过天数</span>
            <span class="summary-value">{{ dailyStats.elapsedDays }} 天 / {{ dailyStats.daysInMonth }} 天</span>
          </div>
          <div v-if="dailyStats.isCurrentMonth && dailyStats.remainingDays > 0" class="summary-row">
            <span class="summary-label">剩余天数</span>
            <span class="summary-value">{{ dailyStats.remainingDays }} 天</span>
          </div>
        </div>
      </article>

      <!-- 本月大额支出 -->
      <article class="stats-module">
        <div class="stats-module-header">
          <h3 class="stats-module-title">本月大额支出</h3>
          <button class="stats-view-all" @click="openInsightView('expenses')">查看全部 &rarr;</button>
        </div>
        <div v-if="largeExpenses.length > 0" class="stats-list">
          <div
            v-for="bill in largeExpenses"
            :key="'le-'+bill.id"
            class="expense-row"
            @click="openBillDetail(bill.id)"
          >
            <span class="expense-date">{{ bill.bill_time.slice(5, 10) }}</span>
            <span class="expense-title">{{ categoryText(bill) }}<template v-if="bill.merchant"> / {{ bill.merchant }}</template></span>
            <span class="expense-amount">{{ formatBillAmount(bill) }}</span>
          </div>
        </div>
        <p v-else class="stats-empty">本月暂无大额支出</p>
      </article>

      <!-- 常去商家 -->
      <article class="stats-module">
        <div class="stats-module-header">
          <h3 class="stats-module-title">常去商家</h3>
          <button class="stats-view-all" @click="openInsightView('merchants')">查看全部 &rarr;</button>
        </div>
        <div v-if="merchantTop5.length > 0" class="stats-list">
          <div v-for="item in merchantTop5" :key="item.merchant" class="merchant-row">
            <span class="merchant-name">{{ item.merchant }}</span>
            <span class="merchant-amount">{{ formatMoney(item.total) }}</span>
            <span class="merchant-count">{{ item.count }} 笔</span>
          </div>
        </div>
        <p v-else class="stats-empty">暂无商家记录</p>
      </article>
    </section>

  </main>

    <!-- Category Management Page -->
    <main class="page-view" v-else-if="currentView === 'categories'">
      <header class="page-header">
        <div>
          <p class="eyebrow">SETTINGS · 分类</p>
          <h1 class="page-title">分类管理</h1>
          <p class="page-desc">维护收入 / 支出的一级和二级分类</p>
        </div>
        <div class="page-header-actions">
          <button type="button" :disabled="categoryManagerLoading" @click="refreshCategoryManagement">刷新</button>
        </div>
      </header>

      <div class="page-toolbar">
        <div class="manager-tabs" aria-label="分类类型">
          <button type="button" :class="{ active: categoryManageType === 'expense' }" @click="categoryManageType = 'expense'">
            支出分类
          </button>
          <button type="button" :class="{ active: categoryManageType === 'income' }" @click="categoryManageType = 'income'">
            收入分类
          </button>
        </div>
        <div class="page-toolbar-right">
          <input v-model="categoryManageSearch" type="search" placeholder="搜索分类..." class="page-search" />
          <button type="button" @click="toggleCategoryEditor">
            {{ categoryEditorOpen ? '收起' : '+ 新增分类' }}
          </button>
        </div>
      </div>

      <section v-if="categoryEditorOpen" class="manager-editor">
        <div class="manager-editor-title">{{ categoryEditingId ? '编辑分类' : '新增分类' }}</div>
        <label>
          <span>层级</span>
          <select v-model.number="categoryForm.parent_id" :disabled="Boolean(categoryEditingId)">
            <option :value="0">一级分类</option>
            <option v-for="parent in categoryParentOptions" :key="parent.id" :value="parent.id">
              {{ parent.name }} 的子分类
            </option>
          </select>
        </label>
        <label>
          <span>名称</span>
          <input v-model="categoryForm.name" type="text" placeholder="分类名称" @keydown.enter="saveManagedCategory" />
        </label>
        <div class="field-label">图标</div>
        <IconPicker v-model="categoryForm.icon_key" />
        <div class="manager-form-actions">
          <button type="button" :disabled="categoryManagerLoading" @click="saveManagedCategory">
            {{ categoryManagerLoading ? '保存中' : (categoryEditingId ? '保存' : '创建') }}
          </button>
          <button type="button" class="btn-ghost" @click="resetCategoryForm()">清空</button>
        </div>
      </section>

      <p v-if="categoryManagerError" class="settings-error">{{ categoryManagerError }}</p>
      <p v-if="categoryManagerMessage" class="settings-message">{{ categoryManagerMessage }}</p>

      <p v-if="categoryManagerLoading && managedCategories.length === 0" class="soft-hint" style="padding:20px 0">正在加载分类...</p>

      <div v-else-if="filteredManagedCategoryRoots.length === 0 && !categoryManagerLoading" class="page-empty">
        <p v-if="categoryManageSearch">没有匹配「{{ categoryManageSearch }}」的分类</p>
        <p v-else>暂无{{ categoryManageType === 'income' ? '收入' : '支出' }}分类，点击上方「+ 新增分类」创建</p>
      </div>

      <div v-else class="manager-list">
        <article
          v-for="category in filteredManagedCategoryRoots"
          :key="category.id"
          class="manager-card"
          :class="{ inactive: !category.is_active, 'drag-over': dragOverCategoryId === category.id && !dragIsChild, 'editing': inlineEditingId === category.id }"
          :draggable="inlineEditingId !== category.id"
          @dragstart="inlineEditingId === category.id ? null : onCategoryDragStart($event, category, false)"
          @dragover="inlineEditingId === category.id ? null : onCategoryDragOver($event, category, false)"
          @dragleave="inlineEditingId === category.id ? null : onCategoryDragLeave"
          @drop="inlineEditingId === category.id ? null : onCategoryDrop($event, category, false)"
          @dragend="inlineEditingId === category.id ? null : onDragEnd"
        >
          <!-- View mode -->
          <div v-if="inlineEditingId !== category.id" class="manager-row">
            <span class="drag-handle" title="拖拽排序">⋮⋮</span>
            <div class="manager-main">
              <strong><AppIcon :name="getCategoryIconKey(category)" :size="20" /> {{ category.name }}</strong>
              <span class="manager-meta">
                {{ category.bill_count || 0 }} 笔
                <template v-if="category.last_used_at"> · 最近 {{ formatShortDateTime(category.last_used_at) }}</template>
              </span>
            </div>
            <div class="manager-actions">
              <button type="button" class="btn-ghost" @click="resetCategoryForm(category)">+ 子分类</button>
              <button type="button" class="btn-ghost" @click="startInlineEdit(category)">编辑</button>
              <button type="button" class="btn-ghost" @click="toggleCategoryActive(category)">
                {{ category.is_active ? '停用' : '启用' }}
              </button>
              <button
                type="button"
                class="btn-ghost"
                :disabled="category.bill_count > 0 || (category.children && category.children.length > 0)"
                @click="deleteManagedCategory(category)"
              >
                删除
              </button>
            </div>
          </div>
          <!-- Edit mode -->
          <div v-else class="manager-row manager-row-edit">
            <span class="drag-handle drag-handle-disabled">⋮⋮</span>
            <div class="manager-main">
              <div class="inline-edit-row">
                <button type="button" class="inline-icon-btn" @click="inlineIconPickerOpen = !inlineIconPickerOpen">
                  <AppIcon :name="inlineDraft.icon_key || 'unknown'" :size="20" />
                </button>
                <input
                  v-model="inlineDraft.name"
                  type="text"
                  class="inline-name-input"
                  @keydown.enter="saveInlineEdit(category)"
                  @keydown.escape="cancelInlineEdit"
                />
              </div>
              <IconPicker v-if="inlineIconPickerOpen" v-model="inlineDraft.icon_key" class="inline-icon-picker" />
              <span class="manager-meta">
                {{ category.bill_count || 0 }} 笔
                <template v-if="category.last_used_at"> · 最近 {{ formatShortDateTime(category.last_used_at) }}</template>
              </span>
            </div>
            <div class="manager-actions">
              <button type="button" class="btn-ghost" :disabled="categoryManagerLoading" @click="saveInlineEdit(category)">保存</button>
              <button type="button" class="btn-ghost" @click="cancelInlineEdit">取消</button>
            </div>
          </div>
          <div v-if="category.children && category.children.length > 0" class="manager-children">
            <div
              v-for="child in category.children"
              :key="child.id"
              class="manager-child-row"
              :class="{ inactive: !child.is_active, 'drag-over': dragOverCategoryId === child.id && dragIsChild, 'editing': inlineEditingId === child.id }"
              :draggable="inlineEditingId !== child.id"
              @dragstart="inlineEditingId === child.id ? null : onCategoryDragStart($event, child, true)"
              @dragover="inlineEditingId === child.id ? null : onCategoryDragOver($event, child, true)"
              @dragleave="inlineEditingId === child.id ? null : onCategoryDragLeave"
              @drop="inlineEditingId === child.id ? null : onCategoryDrop($event, child, true)"
              @dragend="inlineEditingId === child.id ? null : onDragEnd"
            >
              <!-- View mode -->
              <template v-if="inlineEditingId !== child.id">
                <div class="manager-child-name"><span class="drag-handle drag-handle-child" title="拖拽排序">⋮⋮</span><AppIcon :name="getCategoryIconKey(child)" :size="18" /> {{ child.name }}</div>
                <div class="manager-child-count">{{ child.bill_count || 0 }} 笔</div>
                <div class="manager-child-time">{{ child.last_used_at ? formatShortDateTime(child.last_used_at) : '暂无记录' }}</div>
                <div class="manager-actions">
                  <button type="button" class="btn-ghost" @click="startInlineEdit(child)">编辑</button>
                  <button type="button" class="btn-ghost" @click="toggleCategoryActive(child)">
                    {{ child.is_active ? '停用' : '启用' }}
                  </button>
                  <button
                    type="button"
                    class="btn-ghost"
                    :disabled="child.bill_count > 0"
                    @click="deleteManagedCategory(child)"
                  >
                    删除
                  </button>
                </div>
              </template>
              <!-- Edit mode -->
              <template v-else>
                <div class="manager-child-name">
                  <span class="drag-handle drag-handle-child drag-handle-disabled">⋮⋮</span>
                  <button type="button" class="inline-icon-btn inline-icon-btn-sm" @click="inlineIconPickerOpen = !inlineIconPickerOpen">
                    <AppIcon :name="inlineDraft.icon_key || 'unknown'" :size="18" />
                  </button>
                  <input
                    v-model="inlineDraft.name"
                    type="text"
                    class="inline-name-input inline-name-input-sm"
                    @keydown.enter="saveInlineEdit(child)"
                    @keydown.escape="cancelInlineEdit"
                  />
                </div>
                <IconPicker v-if="inlineIconPickerOpen" v-model="inlineDraft.icon_key" class="inline-icon-picker" />
                <div class="manager-child-count">{{ child.bill_count || 0 }} 笔</div>
                <div class="manager-child-time">{{ child.last_used_at ? formatShortDateTime(child.last_used_at) : '暂无记录' }}</div>
                <div class="manager-actions">
                  <button type="button" class="btn-ghost" :disabled="categoryManagerLoading" @click="saveInlineEdit(child)">保存</button>
                  <button type="button" class="btn-ghost" @click="cancelInlineEdit">取消</button>
                </div>
              </template>
            </div>
          </div>
        </article>
      </div>
    </main>

    <!-- Tag Management Page -->
    <main class="page-view" v-else-if="currentView === 'tags'">
      <header class="page-header">
        <div>
          <p class="eyebrow">SETTINGS · 标签</p>
          <h1 class="page-title">标签管理</h1>
          <p class="page-desc">整理账单中的场景、人物、支付方式和补充信息</p>
        </div>
        <div class="page-header-actions">
          <button type="button" :disabled="tagManagerLoading" @click="refreshTagManagement">刷新</button>
        </div>
      </header>

      <div class="page-toolbar">
        <div class="manager-tabs manager-tabs-wrap" aria-label="标签分组">
          <button type="button" :class="{ active: tagManageGroup === 'all' }" @click="tagManageGroup = 'all'">
            全部
          </button>
          <button
            v-for="group in tagGroupOptions"
            :key="group"
            type="button"
            :class="{ active: tagManageGroup === group }"
            @click="tagManageGroup = group"
          >
            {{ groupLabel(group) }}
          </button>
        </div>
        <div class="page-toolbar-right">
          <input v-model="tagManageSearch" type="search" placeholder="搜索标签..." class="page-search" />
          <button
            type="button"
            class="btn-ghost"
            :class="{ 'filter-active': tagShowInactiveOnly }"
            @click="tagShowInactiveOnly = !tagShowInactiveOnly"
          >
            只看停用
          </button>
          <button
            type="button"
            class="btn-ghost"
            :class="{ 'filter-active': tagShowUngroupedOnly }"
            @click="tagShowUngroupedOnly = !tagShowUngroupedOnly"
          >
            只看未分组
          </button>
          <button type="button" @click="toggleTagEditor">
            {{ tagEditorOpen ? '收起' : '+ 新增标签' }}
          </button>
        </div>
      </div>

      <section v-if="tagEditorOpen" class="manager-editor">
        <div class="manager-editor-title">{{ tagEditingId ? '编辑标签' : '新增标签' }}</div>
        <label>
          <span>名称</span>
          <input v-model="tagForm.name" type="text" placeholder="标签名称" @keydown.enter="saveManagedTag" />
        </label>
        <label>
          <span>分组</span>
          <input v-model="tagForm.group_name" type="text" list="tag-group-list-page" placeholder="content" />
          <datalist id="tag-group-list-page">
            <option v-for="group in tagGroupOptions" :key="group" :value="group">{{ groupLabel(group) }}</option>
          </datalist>
        </label>
        <label>
          <span>排序</span>
          <input v-model.number="tagForm.sort_order" type="number" min="0" placeholder="0" />
        </label>
        <div class="manager-form-actions">
          <button type="button" :disabled="tagManagerLoading" @click="saveManagedTag">
            {{ tagManagerLoading ? '保存中' : (tagEditingId ? '保存' : '创建') }}
          </button>
          <button type="button" class="btn-ghost" @click="resetTagForm()">清空</button>
        </div>
      </section>

      <section v-if="mergeSourceTag" class="manager-editor merge-editor">
        <div class="manager-editor-title">合并「{{ mergeSourceTag.name }}」到...</div>
        <label>
          <span>目标标签</span>
          <select v-model="mergeTargetTagId">
            <option value="">选择目标标签</option>
            <option v-for="tag in mergeTargetOptions" :key="tag.id" :value="tag.id">
              {{ tag.name }} · {{ groupLabel(tag.group_name) }}
            </option>
          </select>
        </label>
        <div class="manager-form-actions">
          <button type="button" :disabled="tagManagerLoading" @click="mergeManagedTags">确认合并</button>
          <button type="button" class="btn-ghost" @click="cancelMergeTag">取消</button>
        </div>
      </section>

      <p v-if="tagManagerError" class="settings-error">{{ tagManagerError }}</p>
      <p v-if="tagManagerMessage" class="settings-message">{{ tagManagerMessage }}</p>

      <p v-if="tagManagerLoading && managedTags.length === 0" class="soft-hint" style="padding:20px 0">正在加载标签...</p>

      <div v-else-if="visibleManagedTags.length === 0 && !tagManagerLoading" class="page-empty">
        <p v-if="tagManageSearch">没有匹配「{{ tagManageSearch }}」的标签</p>
        <p v-else>暂无标签，点击上方「+ 新增标签」创建</p>
      </div>

      <div v-else class="manager-list">
        <template v-for="[groupName, tags] in groupedVisibleTags" :key="groupName">
          <div class="manager-group-label">{{ groupLabel(groupName) }}</div>
          <article
            v-for="tag in tags"
            :key="tag.id"
            class="manager-card manager-tag-card"
            :class="{ inactive: !tag.is_active }"
          >
            <div class="manager-row">
              <div class="manager-main">
                <strong>{{ tag.name }}</strong>
                <span class="manager-meta">
                  {{ groupLabel(tag.group_name) }} · {{ tag.use_count || 0 }} 次
                  <template v-if="tag.last_used_at"> · 最近 {{ formatShortDateTime(tag.last_used_at) }}</template>
                </span>
              </div>
              <div class="manager-actions">
                <button type="button" class="btn-ghost" @click="openTagEditorForEdit(tag)">编辑</button>
                <button type="button" class="btn-ghost" @click="beginMergeTag(tag)">合并</button>
                <button type="button" class="btn-ghost" @click="toggleManagedTagActive(tag)">
                  {{ tag.is_active ? '停用' : '启用' }}
                </button>
                <button
                  type="button"
                  class="btn-ghost"
                  :disabled="tag.use_count > 0"
                  @click="deleteManagedTag(tag)"
                >
                  删除
                </button>
              </div>
            </div>
          </article>
        </template>
      </div>
    </main>

    <!-- Merchant Management Page -->
    <main class="page-view" v-else-if="currentView === 'merchants'">
      <header class="page-header">
        <div>
          <p class="eyebrow">SETTINGS · 商家</p>
          <h1 class="page-title">商家管理</h1>
          <p class="page-desc">维护账单中常用商家和消费对象</p>
        </div>
        <div class="page-header-actions">
          <button type="button" :disabled="merchantManagerLoading" @click="refreshMerchantManagement">刷新</button>
        </div>
      </header>

      <div class="page-toolbar">
        <div></div>
        <div class="page-toolbar-right">
          <input v-model="merchantManageSearch" type="search" placeholder="搜索商家..." class="page-search" />
          <button type="button" @click="toggleMerchantEditor">
            {{ merchantEditorOpen ? '收起' : '+ 新增商家' }}
          </button>
        </div>
      </div>

      <section v-if="merchantEditorOpen" class="manager-editor">
        <div class="manager-editor-title">{{ merchantEditingId ? '编辑商家' : '新增商家' }}</div>
        <label>
          <span>名称</span>
          <input v-model="merchantForm.name" type="text" placeholder="商家名称" @keydown.enter="saveManagedMerchant" />
        </label>
        <label>
          <span>排序</span>
          <input v-model.number="merchantForm.sort_order" type="number" min="0" placeholder="0" />
        </label>
        <div class="manager-form-actions" style="grid-column: 1 / -1;">
          <button type="button" :disabled="merchantManagerLoading" @click="saveManagedMerchant">
            {{ merchantManagerLoading ? '保存中' : (merchantEditingId ? '保存' : '创建') }}
          </button>
          <button type="button" class="btn-ghost" @click="resetMerchantForm(); merchantEditorOpen = false">取消</button>
        </div>
      </section>

      <p v-if="merchantManagerError" class="settings-error">{{ merchantManagerError }}</p>
      <p v-if="merchantManagerMessage" class="settings-message">{{ merchantManagerMessage }}</p>

      <p v-if="merchantManagerLoading && managedMerchants.length === 0" class="soft-hint" style="padding:20px 0">正在加载商家...</p>

      <div v-else-if="filteredManagedMerchants.length === 0 && !merchantManagerLoading" class="page-empty">
        <p v-if="merchantManageSearch">没有匹配「{{ merchantManageSearch }}」的商家</p>
        <p v-else>暂无商家，点击上方「+ 新增商家」创建</p>
      </div>

      <div v-else class="manager-list">
        <article
          v-for="merchant in filteredManagedMerchants"
          :key="merchant.id"
          class="manager-card"
          :class="{ inactive: !merchant.is_active }"
        >
          <div class="manager-row">
            <div class="manager-main">
              <strong>{{ merchant.name }}</strong>
              <span class="manager-meta">
                {{ formatMoney(merchant.total_amount) }} · {{ merchant.use_count || 0 }} 笔
                <template v-if="merchant.last_used_at"> · 最近 {{ formatShortDateTime(merchant.last_used_at) }}</template>
              </span>
            </div>
            <div class="manager-actions">
              <button type="button" class="btn-ghost" @click="editMerchant(merchant)">编辑</button>
              <button type="button" class="btn-ghost" @click="toggleManagedMerchantActive(merchant)">
                {{ merchant.is_active ? '停用' : '启用' }}
              </button>
              <button
                type="button"
                class="btn-ghost"
                :disabled="merchant.use_count > 0"
                @click="deleteManagedMerchant(merchant)"
              >
                删除
              </button>
            </div>
          </div>
        </article>
      </div>
    </main>

    </div><!-- .main-workspace -->

    <aside class="right-panel" aria-label="工作面板" @keydown="onDetailPanelKeydown">
    <template v-if="showSettingsPanel">
      <!-- Category Manager Sub-View -->
      <template v-if="settingsView === 'category-manager'">
        <header class="detail-panel-header">
          <div class="detail-header-row">
            <div>
              <p class="eyebrow">Settings · 分类</p>
              <h2>分类管理</h2>
              <p class="detail-header-subtitle">维护收入 / 支出的一级和二级分类</p>
            </div>
            <button type="button" class="icon-button" @click="goBackToSettings"><AppIcon name="back" :size="16" /></button>
          </div>
        </header>
        <div class="detail-body">
          <div class="manager-toolbar">
            <div class="manager-tabs" aria-label="分类类型">
              <button
                type="button"
                :class="{ active: categoryManageType === 'expense' }"
                @click="categoryManageType = 'expense'"
              >
                支出
              </button>
              <button
                type="button"
                :class="{ active: categoryManageType === 'income' }"
                @click="categoryManageType = 'income'"
              >
                收入
              </button>
            </div>
            <button type="button" class="btn-secondary" @click="resetCategoryForm()">新增一级</button>
          </div>

          <section class="manager-editor">
            <div class="manager-editor-title">{{ categoryEditingId ? '编辑分类' : '新增分类' }}</div>
            <label>
              <span>层级</span>
              <select v-model.number="categoryForm.parent_id" :disabled="Boolean(categoryEditingId)">
                <option :value="0">一级分类</option>
                <option v-for="parent in categoryParentOptions" :key="parent.id" :value="parent.id">
                  {{ parent.name }} 下
                </option>
              </select>
            </label>
            <label>
              <span>名称</span>
              <input v-model="categoryForm.name" type="text" placeholder="分类名称" />
            </label>
            <div class="field-label" style="margin-top:4px">图标</div>
            <IconPicker v-model="categoryForm.icon_key" />
            <div class="manager-form-actions">
              <button type="button" :disabled="categoryManagerLoading" @click="saveManagedCategory">
                {{ categoryManagerLoading ? '保存中' : (categoryEditingId ? '保存分类' : '创建分类') }}
              </button>
              <button type="button" class="btn-ghost" @click="resetCategoryForm()">清空</button>
            </div>
          </section>

          <p v-if="categoryManagerError" class="settings-error">{{ categoryManagerError }}</p>
          <p v-if="categoryManagerMessage" class="settings-message">{{ categoryManagerMessage }}</p>
          <p v-if="categoryManagerLoading && managedCategories.length === 0" class="soft-hint">正在加载分类...</p>

          <div class="manager-list">
            <article
              v-for="category in managedCategoryRoots"
              :key="category.id"
              class="manager-card"
              :class="{ inactive: !category.is_active }"
            >
              <div class="manager-row">
                <div class="manager-main">
                  <strong><AppIcon :name="getCategoryIconKey(category)" :size="20" /> {{ category.name }}</strong>
                  <span class="manager-meta">
                    #{{ category.sort_order || 0 }} · {{ category.bill_count || 0 }} 笔
                    <template v-if="category.last_used_at"> · 最近 {{ formatShortDateTime(category.last_used_at) }}</template>
                  </span>
                </div>
                <div class="manager-actions">
                  <button type="button" class="btn-ghost" @click="resetCategoryForm(category)">子类</button>
                  <button type="button" class="btn-ghost" @click="editCategory(category)">编辑</button>
                  <button type="button" class="btn-ghost" @click="toggleCategoryActive(category)">
                    {{ category.is_active ? '停用' : '启用' }}
                  </button>
                  <button
                    type="button"
                    class="btn-ghost"
                    :disabled="category.bill_count > 0 || (category.children && category.children.length > 0)"
                    @click="deleteManagedCategory(category)"
                  >
                    删除
                  </button>
                </div>
              </div>
              <div v-if="category.children && category.children.length > 0" class="manager-children">
                <div
                  v-for="child in category.children"
                  :key="child.id"
                  class="manager-child-row"
                  :class="{ inactive: !child.is_active }"
                >
                  <div class="manager-main">
                    <span><AppIcon :name="getCategoryIconKey(child)" :size="18" /> {{ child.name }}</span>
                    <span class="manager-meta">
                      #{{ child.sort_order || 0 }} · {{ child.bill_count || 0 }} 笔
                      <template v-if="child.last_used_at"> · 最近 {{ formatShortDateTime(child.last_used_at) }}</template>
                    </span>
                  </div>
                  <div class="manager-actions">
                    <button type="button" class="btn-ghost" @click="editCategory(child)">编辑</button>
                    <button type="button" class="btn-ghost" @click="toggleCategoryActive(child)">
                      {{ child.is_active ? '停用' : '启用' }}
                    </button>
                    <button
                      type="button"
                      class="btn-ghost"
                      :disabled="child.bill_count > 0"
                      @click="deleteManagedCategory(child)"
                    >
                      删除
                    </button>
                  </div>
                </div>
              </div>
            </article>
          </div>
        </div>
        <footer class="detail-footer">
          <span class="detail-status">{{ managedCategoryRoots.length }} 个一级分类</span>
          <div class="detail-footer-actions">
            <button type="button" :disabled="categoryManagerLoading" @click="refreshCategoryManagement">刷新</button>
            <button type="button" @click="goBackToSettings">返回设置</button>
          </div>
        </footer>
      </template>

      <!-- Tag Manager Sub-View -->
      <template v-else-if="settingsView === 'tag-manager'">
        <header class="detail-panel-header">
          <div class="detail-header-row">
            <div>
              <p class="eyebrow">Settings · 标签</p>
              <h2>标签管理</h2>
              <p class="detail-header-subtitle">维护标签分组、启停状态和重复标签合并</p>
            </div>
            <button type="button" class="icon-button" @click="goBackToSettings"><AppIcon name="back" :size="16" /></button>
          </div>
        </header>
        <div class="detail-body">
          <div class="manager-toolbar">
            <div class="manager-tabs manager-tabs-wrap" aria-label="标签分组">
              <button type="button" :class="{ active: tagManageGroup === 'all' }" @click="tagManageGroup = 'all'">
                全部
              </button>
              <button
                v-for="group in tagGroupOptions"
                :key="group"
                type="button"
                :class="{ active: tagManageGroup === group }"
                @click="tagManageGroup = group"
              >
                {{ groupLabel(group) }}
              </button>
            </div>
          </div>

          <section class="manager-editor">
            <div class="manager-editor-title">{{ tagEditingId ? '编辑标签' : '新增标签' }}</div>
            <label>
              <span>名称</span>
              <input v-model="tagForm.name" type="text" placeholder="标签名称" />
            </label>
            <label>
              <span>分组</span>
              <input v-model="tagForm.group_name" type="text" list="tag-group-list" placeholder="content" />
              <datalist id="tag-group-list">
                <option v-for="group in tagGroupOptions" :key="group" :value="group">{{ groupLabel(group) }}</option>
              </datalist>
            </label>
            <label>
              <span>排序</span>
              <input v-model.number="tagForm.sort_order" type="number" min="0" placeholder="自动追加" />
            </label>
            <div class="manager-form-actions">
              <button type="button" :disabled="tagManagerLoading" @click="saveManagedTag">
                {{ tagManagerLoading ? '保存中' : (tagEditingId ? '保存标签' : '创建标签') }}
              </button>
              <button type="button" class="btn-ghost" @click="resetTagForm">清空</button>
            </div>
          </section>

          <section v-if="mergeSourceTag" class="manager-editor merge-editor">
            <div class="manager-editor-title">合并「{{ mergeSourceTag.name }}」</div>
            <label>
              <span>合并到</span>
              <select v-model="mergeTargetTagId">
                <option value="">选择目标标签</option>
                <option v-for="tag in mergeTargetOptions" :key="tag.id" :value="tag.id">
                  {{ tag.name }} · {{ groupLabel(tag.group_name) }}
                </option>
              </select>
            </label>
            <div class="manager-form-actions">
              <button type="button" :disabled="tagManagerLoading" @click="mergeManagedTags">确认合并</button>
              <button type="button" class="btn-ghost" @click="cancelMergeTag">取消</button>
            </div>
          </section>

          <p v-if="tagManagerError" class="settings-error">{{ tagManagerError }}</p>
          <p v-if="tagManagerMessage" class="settings-message">{{ tagManagerMessage }}</p>
          <p v-if="tagManagerLoading && managedTags.length === 0" class="soft-hint">正在加载标签...</p>

          <div class="manager-list">
            <article
              v-for="tag in visibleManagedTags"
              :key="tag.id"
              class="manager-card manager-tag-card"
              :class="{ inactive: !tag.is_active }"
            >
              <div class="manager-row">
                <div class="manager-main">
                  <strong>{{ tag.name }}</strong>
                  <span class="manager-meta">
                    {{ groupLabel(tag.group_name) }} · #{{ tag.sort_order || 0 }} · {{ tag.use_count || 0 }} 笔
                    <template v-if="tag.last_used_at"> · 最近 {{ formatShortDateTime(tag.last_used_at) }}</template>
                  </span>
                </div>
                <div class="manager-actions">
                  <button type="button" class="btn-ghost" @click="editManagedTag(tag)">编辑</button>
                  <button type="button" class="btn-ghost" @click="beginMergeTag(tag)">合并</button>
                  <button type="button" class="btn-ghost" @click="toggleManagedTagActive(tag)">
                    {{ tag.is_active ? '停用' : '启用' }}
                  </button>
                  <button
                    type="button"
                    class="btn-ghost"
                    :disabled="tag.use_count > 0"
                    @click="deleteManagedTag(tag)"
                  >
                    删除
                  </button>
                </div>
              </div>
            </article>
          </div>
        </div>
        <footer class="detail-footer">
          <span class="detail-status">{{ visibleManagedTags.length }} 个标签</span>
          <div class="detail-footer-actions">
            <button type="button" :disabled="tagManagerLoading" @click="refreshTagManagement">刷新</button>
            <button type="button" @click="goBackToSettings">返回设置</button>
          </div>
        </footer>
      </template>

      <!-- Deleted Bills Sub-View -->
      <template v-else-if="settingsView === 'deleted-bills'">
        <header class="detail-panel-header">
          <div class="detail-header-row">
            <div>
              <p class="eyebrow">Settings · 数据维护</p>
              <h2>已删除账单</h2>
              <p class="detail-header-subtitle">这些账单不会出现在默认账单流中，可以在这里恢复</p>
            </div>
            <button type="button" class="icon-button" @click="goBackToSettings"><AppIcon name="back" :size="16" /></button>
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
            <button type="button" class="icon-button" @click="closeSettings"><AppIcon name="close" :size="16" /></button>
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
              <button type="button" class="settings-link-item" @click="closeSettings(); navigateTo('categories')">
                <span class="settings-link-label">分类管理</span>
                <span class="settings-link-desc">维护收入 / 支出分类、层级、启停状态</span>
                <span class="settings-link-arrow">→</span>
              </button>
              <button type="button" class="settings-link-item" @click="closeSettings(); navigateTo('tags')">
                <span class="settings-link-label">标签管理</span>
                <span class="settings-link-desc">维护标签分组、使用情况，并合并重复标签</span>
                <span class="settings-link-arrow">→</span>
              </button>
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
    <template v-else-if="insightView">
      <!-- Categories insight -->
      <template v-if="insightView === 'categories'">
        <header class="detail-panel-header">
          <div class="detail-header-row">
            <div>
              <p class="eyebrow">分类详情</p>
              <h2>本月花在哪</h2>
            </div>
            <button type="button" class="icon-button" @click="closeInsightView"><AppIcon name="close" :size="16" /></button>
          </div>
        </header>
        <div class="detail-body">
          <div v-if="allCategories.length > 0" class="stats-list">
            <div v-for="item in allCategories" :key="item.category" class="stats-row stats-row-cat">
              <span class="stats-label">{{ item.category }}</span>
              <span class="stats-value">{{ formatMoney(item.total) }}</span>
              <span class="stats-pct">{{ Math.round(item.pct) }}%</span>
            </div>
          </div>
          <p v-else class="stats-empty">本月暂无支出</p>
        </div>
      </template>
      <!-- Merchants insight -->
      <template v-else-if="insightView === 'merchants'">
        <header class="detail-panel-header">
          <div class="detail-header-row">
            <div>
              <p class="eyebrow">商家详情</p>
              <h2>常去商家</h2>
            </div>
            <button type="button" class="icon-button" @click="closeInsightView"><AppIcon name="close" :size="16" /></button>
          </div>
        </header>
        <div class="detail-body">
          <div v-if="allMerchants.length > 0" class="stats-list">
            <div v-for="item in allMerchants" :key="item.merchant" class="stats-row">
              <span class="stats-label">{{ item.merchant }}</span>
              <span class="stats-value">{{ formatMoney(item.total) }}</span>
              <span class="stats-count">{{ item.count }} 笔</span>
            </div>
          </div>
          <p v-else class="stats-empty">暂无商家记录</p>
        </div>
      </template>
      <!-- Expenses insight -->
      <template v-else-if="insightView === 'expenses'">
        <header class="detail-panel-header">
          <div class="detail-header-row">
            <div>
              <p class="eyebrow">支出详情</p>
              <h2>本月大额支出</h2>
            </div>
            <button type="button" class="icon-button" @click="closeInsightView"><AppIcon name="close" :size="16" /></button>
          </div>
        </header>
        <div class="detail-body">
          <div v-if="allTopExpenses.length > 0" class="stats-list">
            <div
              v-for="bill in allTopExpenses"
              :key="'ae-'+bill.id"
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
        </div>
      </template>
    </template>
    <template v-else-if="detailPanelStatus === 'ledger'">
      <header class="detail-panel-header">
        <div class="detail-header-row">
          <div>
            <p class="eyebrow">Ledger</p>
            <h2>账单流</h2>
          </div>
          <button type="button" class="btn-create" :disabled="loading !== ''" @click="openCreateBill()">+ 记一笔</button>
        </div>
        <div class="ledger-toolbar-row">
          <div class="search-wrapper">
            <span class="search-icon"><AppIcon name="search" :size="14" /></span>
            <input v-model="searchQuery" @input="onSearchInput" type="search" placeholder="搜索全部账单（分类、商家、备注、标签...）" class="bill-search" />
          </div>
        </div>
        <!-- Search status -->
        <div v-if="isSearching" class="search-status-row">
          <span v-if="searchLoading">正在搜索...</span>
          <span v-else-if="searchPerformed">{{ searchedBills.length }} 条结果</span>
          <button type="button" class="btn-ghost" style="font-size:11px" @click="searchQuery = ''; searchedBills = []; searchPerformed = false">清空</button>
        </div>
        <!-- Filter chips: hidden during search -->
        <div v-if="!isSearching" class="ledger-filter-row">
          <div class="filter-chips">
            <button v-for="opt in [{v:'all',l:'全部'},{v:'expense',l:'支出'},{v:'income',l:'收入'}]" :key="opt.v" class="filter-chip" :class="{ active: filterType === opt.v }" @click="filterType = opt.v">{{ opt.l }}</button>
            <span class="filter-sep"></span>
            <button v-for="cat in visibleFilterCategories" :key="'rp-fc-'+cat" class="filter-chip" :class="{ active: filterCategory === cat }" @click="toggleFilterCategory(cat)">{{ cat }}</button>
            <button v-if="hasMoreFilterCategories" class="filter-chip filter-chip-more" @click="showMoreFilters = !showMoreFilters">{{ showMoreFilters ? '收起' : '更多' }}</button>
            <template v-if="availableFilterTags.length > 0">
              <span class="filter-sep"></span>
              <button
                v-for="tag in availableFilterTags"
                :key="'rp-ft-'+tag"
                class="filter-chip filter-chip-tag"
                :class="{ active: filterTag === tag }"
                @click="toggleFilterTag(tag)"
              >
                #{{ tag }}
              </button>
            </template>
          </div>
          <button v-if="hasActiveFilters" class="btn-clear-filters" @click="clearFilters">清空</button>
        </div>
      </header>
      <div class="detail-body">
        <div v-if="billGroups.length === 0 && !hasActiveFilters" class="empty-ledger" style="margin-top:40px">
          <p class="empty-ledger-title">暂无账单</p>
          <p class="empty-ledger-desc">这个月还没有记录</p>
          <button type="button" class="btn-create" :disabled="loading !== ''" @click="openCreateBill()">+ 记一笔</button>
        </div>
        <div v-else-if="billGroups.length === 0 && hasActiveFilters" class="empty-ledger" style="margin-top:40px">
          <p class="empty-ledger-title">没有匹配账单</p>
          <button type="button" class="btn-clear-filters" @click="clearFilters">清空筛选</button>
        </div>
        <div v-else class="timeline-list">
          <section v-for="group in billGroups" :key="group.date" class="tl-group">
            <div class="tl-date-row">
              <span class="tl-dot"></span>
              <span class="tl-date">{{ group.label }}</span>
              <span class="tl-summary">
                <span v-if="group.expense > 0" class="tl-expense">支出 {{ formatMoney(group.expense) }}</span>
                <span v-if="group.expense > 0 && group.income > 0" class="tl-sep">·</span>
                <span v-if="group.income > 0" class="tl-income">收入 {{ formatMoney(group.income) }}</span>
                <span class="tl-count">{{ group.bills.length }} 笔</span>
              </span>
            </div>
            <div class="tl-group-inner">
              <div class="tl-items">
                <article v-for="bill in group.bills" :key="bill.id" class="tl-item" :class="{ selected: selectedBillId === bill.id, deleted: showDeleted }" @click="openBillDetail(bill.id)">
                  <span class="tl-item-dot"></span>
                  <div class="tl-item-body">
                    <div class="tl-item-row1">
                      <span class="tl-category">{{ categoryText(bill) }}</span>
                      <span class="tl-amount" :class="amountClass(bill.type)">{{ formatBillAmount(bill) }}</span>
                    </div>
                    <div v-if="bill.merchant || bill.note" class="tl-item-row2">{{ [bill.merchant, bill.note].filter(Boolean).join(' · ') }}</div>
                    <div v-if="listTags(bill).length > 0" class="tl-item-row3">
                      <span v-for="tag in listTags(bill).slice(0, 3)" :key="tag" class="tag-chip">{{ tag }}</span>
                    </div>
                  </div>
                </article>
              </div>
            </div>
          </section>
        </div>
      </div>
    </template>
    <template v-else>
      <header class="detail-panel-header">
      <template v-if="detailPanelStatus === 'loading'">
        <div class="detail-header-row">
          <div>
            <p class="eyebrow">Bill detail</p>
            <h2>正在加载账单详情...</h2>
          </div>
          <button type="button" class="icon-button" @click="closeDetailPanel"><AppIcon name="close" :size="16" /></button>
        </div>
      </template>
      <template v-else-if="detailPanelStatus === 'create'">
        <div class="detail-header-row">
          <div>
            <p class="eyebrow">New bill</p>
            <h2>新增账单</h2>
          </div>
          <button type="button" class="icon-button" @click="closeDetailPanel"><AppIcon name="close" :size="16" /></button>
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
          <button type="button" class="icon-button" @click="closeDetailPanel"><AppIcon name="close" :size="16" /></button>
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
            <span v-for="tag in detailForm.tags" :key="tag" class="tag-chip editable-chip selected-tag-chip">
              {{ tag }}
              <button type="button" @click="removeTag(tag)">x</button>
            </span>
          </div>

          <div class="tag-search-row">
            <input v-model="tagSearch" class="tag-search" type="search" placeholder="搜索或添加标签..." @keydown.enter.prevent="createTagFromSearch" />
            <button v-if="showCreateTag" type="button" class="tag-create-btn" @click="createTagFromSearch">
              创建「{{ tagSearchText }}」
            </button>
          </div>

          <div v-if="!tagSearchQuery">
            <div v-if="recentTags.length > 0" class="tag-section">
              <div class="tag-section-label">最近使用</div>
              <div class="chip-picker">
                <button
                  v-for="tag in recentTags" :key="'rt-'+tag" type="button"
                  class="choice-chip"
                  @click="addTag(tag)"
                >{{ tag }}</button>
              </div>
            </div>

            <div v-if="allAvailableTags.length > 0" class="tag-section">
              <div v-if="recentTags.length > 0" class="tag-section-label">全部标签</div>
              <div class="chip-picker">
                <button
                  v-for="tag in allAvailableTags" :key="'at-'+tag" type="button"
                  class="choice-chip"
                  @click="addTag(tag)"
                >{{ tag }}</button>
              </div>
            </div>

            <p v-if="recentTags.length === 0 && allAvailableTags.length === 0" class="soft-hint">暂无可用标签</p>
          </div>

          <div v-else>
            <div v-if="filteredRecentTags.length > 0" class="tag-section">
              <div class="tag-section-label">最近使用</div>
              <div class="chip-picker">
                <button
                  v-for="tag in filteredRecentTags" :key="'frt-'+tag" type="button"
                  class="choice-chip"
                  @click="addTag(tag)"
                >{{ tag }}</button>
              </div>
            </div>

            <div v-if="filteredAvailableTags.length > 0" class="tag-section">
              <div class="tag-section-label">全部标签</div>
              <div class="chip-picker">
                <button
                  v-for="tag in filteredAvailableTags" :key="'fat-'+tag" type="button"
                  class="choice-chip"
                  @click="addTag(tag)"
                >{{ tag }}</button>
              </div>
            </div>

            <p v-if="filteredRecentTags.length === 0 && filteredAvailableTags.length === 0 && !showCreateTag" class="soft-hint">没有匹配标签</p>
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
  </div><!-- .app-shell -->
</template>
