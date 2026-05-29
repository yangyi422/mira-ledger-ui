// ── Icon key → local SVG import mapping ──
// All SVGs in src/assets/icons/*.svg loaded eagerly as raw strings.
// SVGs use stroke="currentColor" and inherit color from parent.

const iconModules = import.meta.glob('../assets/icons/*.svg', {
  eager: true,
  query: '?raw',
  import: 'default',
})

function fileName(path) {
  return path.split('/').pop().replace(/\.svg$/, '')
}

const FILE_MAP = {}
for (const [path, svg] of Object.entries(iconModules)) {
  FILE_MAP[fileName(path)] = svg
}

const FALLBACK_SVG = FILE_MAP['unknown'] || ''

// ── Alias map: iconKey → actual SVG filename ──
// Used when conceptual iconKey differs from the SVG filename on disk.
const ALIAS_MAP = {
  enable: 'check',
  refresh: 'refresh',
  merge: 'edit',
  clear: 'close',
  collapse: 'sort',
  back: 'sort',
  more: 'more',
  save: 'save',

  snack: 'snacks',
  tea: 'coffee',
  grocery: 'grocery',

  kitchen: 'kitchen',
  bathroom: 'cleaning',
  living: 'living',
  rent: 'rent',
  utility: 'electricity',

  repair: 'tools',

  transport: 'car',
  commute: 'metro',
  parking: 'parking',
  travel: 'plane',

  wedding: 'party',
  friend: 'relationship',

  medical: 'health',
  fitness: 'fitness',
  hospital: 'health',
  medicine: 'medicine',

  education: 'education',
  course: 'education',

  entertainment: 'game',
  travel_leisure: 'plane',

  clothing: 'clothes',
  shoes: 'clothes',
  accessory_fashion: 'accessory',

  bonus: 'bonus',
  business: 'briefcase',
  investment: 'investment',
  savings: 'savings',
  loan: 'loan',
  insurance: 'insurance',
  bank: 'bank',
  cash: 'cash',
  wallet: 'wallet',
  'credit-card': 'card',

  phone: 'phone',
  laptop: 'laptop',
  pc: 'laptop',
  camera: 'camera',
  internet: 'internet',
  security: 'security',
  music: 'music',
  movie: 'movie',
  video: 'movie',
  software: 'software',

  car: 'car',
  taxi: 'taxi',
  bus: 'bus',
  metro: 'metro',
  bike: 'bike',
  plane: 'plane',
  train: 'train',
  hotel: 'hotel',
  fuel: 'fuel',
  location: 'location',

  children: 'baby',
  beauty: 'beauty',
  party: 'party',
  gift: 'gift',
  clothes: 'clothes',
  shopping: 'clothes',
  book: 'book',

  electricity: 'electricity',
  gas: 'gas',
  key: 'home',
  note: 'note',
  file: 'file',
  star: 'star',
  clock: 'clock',
  bell: 'bell',
  pet: 'pet',
  family: 'family',
  love: 'love',
  fruit: 'fruit',
  vegetable: 'vegetable',
  drink: 'drink',
  dessert: 'dessert',
  image: 'image',
  sort: 'sort',
  drag: 'drag',
  card: 'card',
}

/**
 * Get raw SVG string for an icon key.
 * Resolution: 1. FILE_MAP direct  2. ALIAS_MAP  3. unknown fallback
 */
export function getIconSvg(key) {
  if (!key) return FALLBACK_SVG
  if (FILE_MAP[key]) return FILE_MAP[key]
  const alias = ALIAS_MAP[key]
  if (alias && FILE_MAP[alias]) return FILE_MAP[alias]
  return FALLBACK_SVG
}

// ── iconGroups: data source for icon picker ──
export const iconGroups = [
  {
    key: 'navigation',
    label: '导航',
    icons: ['overview', 'ledger', 'calendar', 'reports', 'category', 'merchant', 'tag', 'settings', 'note', 'file', 'book', 'image'],
  },
  {
    key: 'actions',
    label: '操作',
    icons: ['add', 'edit', 'delete', 'save', 'check', 'close', 'search', 'filter', 'refresh', 'disable', 'more', 'sort', 'drag', 'star', 'clock', 'bell'],
  },
  {
    key: 'finance',
    label: '财务',
    icons: ['income', 'expense', 'balance', 'wallet', 'card', 'bank', 'cash', 'salary', 'refund', 'subscription', 'investment', 'savings', 'loan', 'insurance', 'invoice', 'receipt', 'budget', 'bonus'],
  },
  {
    key: 'home',
    label: '家居',
    icons: ['home', 'living', 'rent', 'electricity', 'gas', 'cleaning', 'kitchen', 'appliance', 'tools', 'decoration', 'pet'],
  },
  {
    key: 'food',
    label: '餐饮',
    icons: ['dining', 'grocery', 'takeaway', 'breakfast', 'lunch', 'dinner', 'snacks', 'coffee', 'drink', 'dessert', 'fruit', 'vegetable'],
  },
  {
    key: 'life',
    label: '生活',
    icons: ['relationship', 'love', 'family', 'baby', 'education', 'book', 'health', 'medicine', 'fitness', 'beauty', 'clothes', 'gift', 'party', 'game', 'pet'],
  },
  {
    key: 'transport',
    label: '出行',
    icons: ['car', 'taxi', 'bus', 'metro', 'bike', 'plane', 'train', 'hotel', 'fuel', 'parking', 'location', 'travel'],
  },
  {
    key: 'digital',
    label: '数码',
    icons: ['digital', 'hardware', 'accessory', 'phone', 'laptop', 'camera', 'game', 'music', 'movie', 'internet', 'software', 'security'],
  },
]

// ── Chinese category name → icon key fallback ──
const CATEGORY_NAME_TO_ICON = {
  '餐饮': 'dining',
  '早餐': 'breakfast',
  '午餐': 'lunch',
  '晚餐': 'dinner',
  '食材烹饪': 'grocery',
  '外卖点单': 'takeaway',
  '咖啡': 'coffee',
  '奶茶': 'tea',
  '零食': 'snacks',
  '甜点': 'dessert',
  '饮品': 'drink',
  '水果': 'fruit',
  '蔬菜': 'vegetable',

  '居住': 'home',
  '房租': 'rent',
  '居家': 'living',
  '厨房': 'kitchen',
  '卫浴': 'bathroom',
  '清洁': 'cleaning',
  '家电': 'appliance',
  '工具': 'tools',
  '装修装饰': 'decoration',
  '水电燃气': 'utility',

  '数码': 'digital',
  '设备硬件': 'hardware',
  '配件耗材': 'accessory',
  '软件订阅': 'subscription',
  '游戏': 'game',
  '维修': 'repair',

  '交通': 'transport',
  '旅行': 'travel',
  '通勤': 'commute',
  '停车': 'parking',

  '人情关系': 'relationship',
  '恋爱': 'love',
  '礼物': 'gift',
  '婚礼': 'wedding',
  '社交': 'friend',

  '健康': 'health',
  '医疗': 'medical',
  '健身': 'fitness',

  '学习': 'education',
  '书籍': 'book',
  '课程': 'course',

  '娱乐休闲': 'entertainment',
  '电影': 'movie',
  '休闲旅行': 'travel_leisure',

  '服饰': 'clothing',
  '鞋履': 'shoes',
  '配饰': 'accessory_fashion',

  '工资': 'salary',
  '退款': 'refund',
  '奖金': 'bonus',
  '生意': 'business',

  '宠物': 'pet',
  '母婴': 'baby',
  '美容': 'beauty',
  '家庭': 'family',
  '未分类': 'unknown',
}

/**
 * Resolve icon key from category name when icon_key is missing.
 */
export function resolveCategoryIconKey(categoryName, iconKey) {
  if (iconKey) return iconKey
  if (categoryName && CATEGORY_NAME_TO_ICON[categoryName]) {
    return CATEGORY_NAME_TO_ICON[categoryName]
  }
  return 'unknown'
}

export function getCategoryIconKey(category) {
  if (!category) return 'unknown'
  if (category.icon_key) return category.icon_key
  if (category.name && CATEGORY_NAME_TO_ICON[category.name]) {
    return CATEGORY_NAME_TO_ICON[category.name]
  }
  return 'unknown'
}

// ── iconKey → Chinese label ──
export const iconLabels = {
  // 导航
  overview: '概览',
  ledger: '账本',
  calendar: '日历',
  reports: '报表',
  category: '分类',
  merchant: '商家',
  tag: '标签',
  settings: '设置',
  note: '笔记',
  file: '文件',
  book: '书籍',
  image: '图片',

  // 操作
  add: '新增',
  edit: '编辑',
  delete: '删除',
  save: '保存',
  check: '确认',
  close: '关闭',
  search: '搜索',
  filter: '筛选',
  refresh: '刷新',
  disable: '停用',
  more: '更多',
  sort: '排序',
  drag: '拖拽',
  star: '收藏',
  clock: '时间',
  bell: '通知',

  // 财务
  income: '收入',
  expense: '支出',
  balance: '结余',
  wallet: '钱包',
  card: '银行卡',
  bank: '银行',
  cash: '现金',
  salary: '工资',
  refund: '退款',
  subscription: '订阅',
  investment: '投资',
  savings: '储蓄',
  loan: '贷款',
  insurance: '保险',
  invoice: '发票',
  receipt: '收据',
  budget: '预算',
  bonus: '奖金',

  // 家居
  home: '家居',
  living: '客厅',
  rent: '房租',
  electricity: '电费',
  gas: '燃气',
  cleaning: '清洁',
  kitchen: '厨房',
  appliance: '家电',
  tools: '工具',
  decoration: '装饰',
  pet: '宠物',

  // 餐饮
  dining: '餐饮',
  grocery: '食材',
  takeaway: '外卖',
  breakfast: '早餐',
  lunch: '午餐',
  dinner: '晚餐',
  snacks: '零食',
  coffee: '咖啡',
  drink: '饮品',
  dessert: '甜点',
  fruit: '水果',
  vegetable: '蔬菜',

  // 生活
  relationship: '人情',
  love: '恋爱',
  family: '家庭',
  baby: '母婴',
  education: '教育',
  health: '健康',
  medicine: '药品',
  fitness: '健身',
  beauty: '美容',
  clothes: '服饰',
  gift: '礼物',
  party: '聚会',
  game: '娱乐',

  // 出行
  car: '汽车',
  taxi: '出租车',
  bus: '公交',
  metro: '地铁',
  bike: '单车',
  plane: '飞机',
  train: '火车',
  hotel: '酒店',
  fuel: '加油',
  parking: '停车',
  location: '位置',
  travel: '旅行',

  // 数码
  digital: '数码',
  hardware: '硬件',
  accessory: '配件',
  phone: '手机',
  laptop: '笔记本',
  camera: '相机',
  music: '音乐',
  movie: '电影',
  internet: '网络',
  software: '软件',
  security: '安全',

  // 其他
  unknown: '未知',
}

export function getIconLabel(key) {
  if (!key) return '未知'
  return iconLabels[key] || key
}

export const ICON_DEFAULTS = { size: 16 }
