export interface SecurePayload {
  version: number
  secure_id: string
  document_hash: string
  encrypted_payload: string
  outer_signature: string
  timestamp: number
  valid_from: number
  valid_until: number
  public_key: string
}

export interface VerificationResult {
  status: 'authentic' | 'tampered' | 'not_yet_valid' | 'expired' | 'error'
  message: string
  document_hash: string
  file_name: string
  file_size: number
  issuer_id: string
  issued_at: string
  valid_from: string
  valid_until: string
  metadata: string
  public_key_hex: string
  scan_count: number
}

export interface KeyPairInfo {
  id: string
  name: string
  public_key: string
  fingerprint: string
  created_at: string
  is_default: boolean
  has_private_key: boolean
}

export interface IssuerConfig {
  file_path: string
  key_pair_id: string
  valid_from: string
  valid_until: string
  metadata: string
  issuer_id: string
  qr_position: string
  qr_page: number
  qr_size: number
}

export interface QRGenerationResult {
  success: boolean
  qr_code_base64: string
  secure_id: string
  document_hash: string
  signed_file_path?: string
  is_pdf: boolean
  error_message?: string
}

export interface DocumentInfo {
  name: string
  size: number
  path: string
}
