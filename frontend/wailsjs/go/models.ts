export namespace model {
	
	export class IssuerConfig {
	    file_path: string;
	    key_pair_id: string;
	    valid_from: string;
	    valid_until: string;
	    metadata: string;
	    issuer_id: string;
	    qr_position: string;
	    qr_page: number;
	    qr_size: number;
	
	    static createFrom(source: any = {}) {
	        return new IssuerConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.file_path = source["file_path"];
	        this.key_pair_id = source["key_pair_id"];
	        this.valid_from = source["valid_from"];
	        this.valid_until = source["valid_until"];
	        this.metadata = source["metadata"];
	        this.issuer_id = source["issuer_id"];
	        this.qr_position = source["qr_position"];
	        this.qr_page = source["qr_page"];
	        this.qr_size = source["qr_size"];
	    }
	}
	export class KeyPairInfo {
	    id: string;
	    name: string;
	    public_key: string;
	    fingerprint: string;
	    created_at: string;
	    is_default: boolean;
	    has_private_key: boolean;
	
	    static createFrom(source: any = {}) {
	        return new KeyPairInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.public_key = source["public_key"];
	        this.fingerprint = source["fingerprint"];
	        this.created_at = source["created_at"];
	        this.is_default = source["is_default"];
	        this.has_private_key = source["has_private_key"];
	    }
	}
	export class QRGenerationResult {
	    success: boolean;
	    qr_code_base64: string;
	    secure_id: string;
	    document_hash: string;
	    signed_file_path?: string;
	    is_pdf: boolean;
	    error_message?: string;
	
	    static createFrom(source: any = {}) {
	        return new QRGenerationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.qr_code_base64 = source["qr_code_base64"];
	        this.secure_id = source["secure_id"];
	        this.document_hash = source["document_hash"];
	        this.signed_file_path = source["signed_file_path"];
	        this.is_pdf = source["is_pdf"];
	        this.error_message = source["error_message"];
	    }
	}
	export class VerificationResult {
	    status: string;
	    message: string;
	    document_hash: string;
	    file_name: string;
	    file_size: number;
	    issuer_id: string;
	    issued_at: string;
	    valid_from: string;
	    valid_until: string;
	    metadata: string;
	    public_key_hex: string;
	    scan_count: number;
	
	    static createFrom(source: any = {}) {
	        return new VerificationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.message = source["message"];
	        this.document_hash = source["document_hash"];
	        this.file_name = source["file_name"];
	        this.file_size = source["file_size"];
	        this.issuer_id = source["issuer_id"];
	        this.issued_at = source["issued_at"];
	        this.valid_from = source["valid_from"];
	        this.valid_until = source["valid_until"];
	        this.metadata = source["metadata"];
	        this.public_key_hex = source["public_key_hex"];
	        this.scan_count = source["scan_count"];
	    }
	}

}

