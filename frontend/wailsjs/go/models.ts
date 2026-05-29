export namespace main {
	
	export class AppConfig {
	    account_book_exe: string;
	    db_path: string;
	    backup_repo: string;
	    default_month: string;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.account_book_exe = source["account_book_exe"];
	        this.db_path = source["db_path"];
	        this.backup_repo = source["backup_repo"];
	        this.default_month = source["default_month"];
	    }
	}
	export class BillDetail {
	    id: number;
	    bill_time: string;
	    type: string;
	    amount: number;
	    category: string;
	    sub_category: string;
	    merchant: string;
	    tags: string[];
	    note: string;
	    display_title: string;
	    display_subtitle: string;
	    raw_category: string;
	    raw_sub_category: string;
	    raw_tags: string[];
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new BillDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.bill_time = source["bill_time"];
	        this.type = source["type"];
	        this.amount = source["amount"];
	        this.category = source["category"];
	        this.sub_category = source["sub_category"];
	        this.merchant = source["merchant"];
	        this.tags = source["tags"];
	        this.note = source["note"];
	        this.display_title = source["display_title"];
	        this.display_subtitle = source["display_subtitle"];
	        this.raw_category = source["raw_category"];
	        this.raw_sub_category = source["raw_sub_category"];
	        this.raw_tags = source["raw_tags"];
	        this.created_at = source["created_at"];
	    }
	}
	export class BillItem {
	    id: number;
	    bill_time: string;
	    type: string;
	    amount: number;
	    category: string;
	    sub_category: string;
	    merchant: string;
	    tags: string[];
	    note: string;
	    display_title: string;
	    display_subtitle: string;
	
	    static createFrom(source: any = {}) {
	        return new BillItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.bill_time = source["bill_time"];
	        this.type = source["type"];
	        this.amount = source["amount"];
	        this.category = source["category"];
	        this.sub_category = source["sub_category"];
	        this.merchant = source["merchant"];
	        this.tags = source["tags"];
	        this.note = source["note"];
	        this.display_title = source["display_title"];
	        this.display_subtitle = source["display_subtitle"];
	    }
	}
	export class CategoryInput {
	    parent_id: number;
	    name: string;
	    type: string;
	    sort_order: number;
	    icon_key: string;
	
	    static createFrom(source: any = {}) {
	        return new CategoryInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.parent_id = source["parent_id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.sort_order = source["sort_order"];
	        this.icon_key = source["icon_key"];
	    }
	}
	export class CategoryNode {
	    id: number;
	    parent_id: number;
	    name: string;
	    type: string;
	    taxonomy_version: string;
	    sort_order: number;
	    is_active: boolean;
	    icon_key: string;
	    bill_count: number;
	    last_used_at: string;
	    children: CategoryNode[];
	
	    static createFrom(source: any = {}) {
	        return new CategoryNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.parent_id = source["parent_id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.taxonomy_version = source["taxonomy_version"];
	        this.sort_order = source["sort_order"];
	        this.is_active = source["is_active"];
	        this.icon_key = source["icon_key"];
	        this.bill_count = source["bill_count"];
	        this.last_used_at = source["last_used_at"];
	        this.children = this.convertValues(source["children"], CategoryNode);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ConfigStatus {
	    account_book_exe: string;
	    account_book_exe_exists: boolean;
	    db_path: string;
	    db_path_exists: boolean;
	    backup_repo: string;
	    backup_repo_exists: boolean;
	    default_month: string;
	    overall_status: string;
	    errors: string[];
	
	    static createFrom(source: any = {}) {
	        return new ConfigStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.account_book_exe = source["account_book_exe"];
	        this.account_book_exe_exists = source["account_book_exe_exists"];
	        this.db_path = source["db_path"];
	        this.db_path_exists = source["db_path_exists"];
	        this.backup_repo = source["backup_repo"];
	        this.backup_repo_exists = source["backup_repo_exists"];
	        this.default_month = source["default_month"];
	        this.overall_status = source["overall_status"];
	        this.errors = source["errors"];
	    }
	}
	export class CreateBillInput {
	    bill_time: string;
	    type: string;
	    amount: number;
	    category: string;
	    sub_category: string;
	    merchant: string;
	    tags: string[];
	    note: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateBillInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bill_time = source["bill_time"];
	        this.type = source["type"];
	        this.amount = source["amount"];
	        this.category = source["category"];
	        this.sub_category = source["sub_category"];
	        this.merchant = source["merchant"];
	        this.tags = source["tags"];
	        this.note = source["note"];
	    }
	}
	export class DashboardStats {
	    month: string;
	    income: number;
	    expense: number;
	    balance: number;
	    refund: number;
	    reimbursement: number;
	    need_review_count: number;
	    total_bills: number;
	    active_bills: number;
	    deleted_bills: number;
	    last_backup_time: string;
	
	    static createFrom(source: any = {}) {
	        return new DashboardStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.month = source["month"];
	        this.income = source["income"];
	        this.expense = source["expense"];
	        this.balance = source["balance"];
	        this.refund = source["refund"];
	        this.reimbursement = source["reimbursement"];
	        this.need_review_count = source["need_review_count"];
	        this.total_bills = source["total_bills"];
	        this.active_bills = source["active_bills"];
	        this.deleted_bills = source["deleted_bills"];
	        this.last_backup_time = source["last_backup_time"];
	    }
	}
	export class MerchantInput {
	    name: string;
	    sort_order: number;
	
	    static createFrom(source: any = {}) {
	        return new MerchantInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.sort_order = source["sort_order"];
	    }
	}
	export class MerchantItem {
	    id: number;
	    name: string;
	    sort_order: number;
	    is_active: boolean;
	    use_count: number;
	    total_amount: number;
	    last_used_at: string;
	
	    static createFrom(source: any = {}) {
	        return new MerchantItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.sort_order = source["sort_order"];
	        this.is_active = source["is_active"];
	        this.use_count = source["use_count"];
	        this.total_amount = source["total_amount"];
	        this.last_used_at = source["last_used_at"];
	    }
	}
	export class MergeTagsInput {
	    source_id: number;
	    target_id: number;
	
	    static createFrom(source: any = {}) {
	        return new MergeTagsInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.source_id = source["source_id"];
	        this.target_id = source["target_id"];
	    }
	}
	export class TagInput {
	    name: string;
	    group_name: string;
	    sort_order: number;
	
	    static createFrom(source: any = {}) {
	        return new TagInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.group_name = source["group_name"];
	        this.sort_order = source["sort_order"];
	    }
	}
	export class TagItem {
	    id: number;
	    name: string;
	    group_name: string;
	    sort_order: number;
	    is_active: boolean;
	    use_count: number;
	    last_used_at: string;
	
	    static createFrom(source: any = {}) {
	        return new TagItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.group_name = source["group_name"];
	        this.sort_order = source["sort_order"];
	        this.is_active = source["is_active"];
	        this.use_count = source["use_count"];
	        this.last_used_at = source["last_used_at"];
	    }
	}
	export class UpdateBillBasicInput {
	    id: number;
	    bill_time: string;
	    type: string;
	    amount: number;
	    merchant: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateBillBasicInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.bill_time = source["bill_time"];
	        this.type = source["type"];
	        this.amount = source["amount"];
	        this.merchant = source["merchant"];
	    }
	}
	export class UpdateCategoryInput {
	    id: number;
	    name: string;
	    sort_order: number;
	    icon_key: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateCategoryInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.sort_order = source["sort_order"];
	        this.icon_key = source["icon_key"];
	    }
	}
	export class UpdateMerchantInput {
	    id: number;
	    name: string;
	    sort_order: number;
	
	    static createFrom(source: any = {}) {
	        return new UpdateMerchantInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.sort_order = source["sort_order"];
	    }
	}
	export class UpdateTagInput {
	    id: number;
	    name: string;
	    group_name: string;
	    sort_order: number;
	
	    static createFrom(source: any = {}) {
	        return new UpdateTagInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.group_name = source["group_name"];
	        this.sort_order = source["sort_order"];
	    }
	}

}

