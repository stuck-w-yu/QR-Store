/**
 * Google Drive Image Bucket Client Service
 * 
 * Menangani upload dan penghapusan gambar di Google Drive melalui Google Apps Script Web App.
 * Dilengkapi dengan kompresi gambar berbasis canvas di sisi klien agar upload cepat
 * dan hemat kuota data.
 */

export const GDRIVE_UPLOAD_URL: string =
	(typeof import.meta !== 'undefined' && import.meta.env?.VITE_GDRIVE_UPLOAD_URL) ||
	'https://script.google.com/macros/s/AKfycbyU1JH6bhyUALSodmzt4vyqneMx3_vI0V7WxS7VCiAJGZ904JVcgg8_YVulTrn5lD8jnA/exec';

export interface GDriveUploadOptions {
	filename?: string;
	folder?: string;
	compress?: boolean;
	maxWidth?: number;
	maxHeight?: number;
	quality?: number;
}

export interface GDriveUploadResult {
	status: 'success' | 'error';
	fileId?: string;
	fileName?: string;
	fileUrl?: string;       // Link preview Google Drive
	directUrl?: string;     // Link CDN langsung untuk <img> HTML (lh3.googleusercontent.com)
	thumbnailUrl?: string;
	size?: number;
	mimeType?: string;
	folder?: string;
	message?: string;
}

/**
 * Mengekstrak Google Drive File ID dari berbagai format URL atau string ID
 */
export function extractGDriveFileId(input?: string | null): string | null {
	if (!input || typeof input !== 'string') return null;
	const trimmed = input.trim();

	// Pola link CDN atau viewer: .../d/FILE_ID
	const dMatch = trimmed.match(/\/d\/([a-zA-Z0-9_-]+)/);
	if (dMatch && dMatch[1]) return dMatch[1];

	// Pola parameter query: ?id=FILE_ID atau &id=FILE_ID
	const idMatch = trimmed.match(/[?&]id=([a-zA-Z0-9_-]+)/);
	if (idMatch && idMatch[1]) return idMatch[1];

	// Pola ID murni (minimal 15 karakter alfanumerik)
	if (/^[a-zA-Z0-9_-]{15,}$/.test(trimmed)) {
		return trimmed;
	}

	return null;
}

/**
 * Memeriksa apakah suatu URL berasal dari Google Drive atau upload GDrive
 */
export function isGDriveUrl(url?: string | null): boolean {
	if (!url || typeof url !== 'string') return false;
	return (
		url.includes('googleusercontent.com') ||
		url.includes('drive.google.com') ||
		url.includes('script.google.com')
	);
}

/**
 * Kompres gambar menggunakan HTML5 Canvas sebelum diunggah
 * Menghasilkan data URL JPEG teroptimasi
 */
export async function compressImage(
	file: File | Blob,
	maxWidth = 1200,
	maxHeight = 1200,
	quality = 0.8
): Promise<{ base64: string; mimeType: string }> {
	return new Promise((resolve, reject) => {
		if (typeof window === 'undefined') {
			return reject(new Error('Canvas kompresi hanya dapat dijalankan di browser.'));
		}

		const reader = new FileReader();
		reader.onerror = () => reject(new Error('Gagal membaca file'));
		reader.onload = (e) => {
			const img = new Image();
			img.onerror = () => reject(new Error('Gagal memproses gambar untuk kompresi'));
			img.onload = () => {
				let width = img.width;
				let height = img.height;

				// Hitung rasio dimensi baru
				if (width > height) {
					if (width > maxWidth) {
						height = Math.round((height * maxWidth) / width);
						width = maxWidth;
					}
				} else {
					if (height > maxHeight) {
						width = Math.round((width * maxHeight) / height);
						height = maxHeight;
					}
				}

				const canvas = document.createElement('canvas');
				canvas.width = width;
				canvas.height = height;

				const ctx = canvas.getContext('2d');
				if (!ctx) {
					return resolve({ base64: e.target?.result as string, mimeType: file.type || 'image/jpeg' });
				}

				// Render gambar ke canvas
				ctx.drawImage(img, 0, 0, width, height);

				// Ekspor sebagai JPEG terkompresi
				const compressedDataUrl = canvas.toDataURL('image/jpeg', quality);
				resolve({ base64: compressedDataUrl, mimeType: 'image/jpeg' });
			};

			img.src = e.target?.result as string;
		};

		reader.readAsDataURL(file);
	});
}

/**
 * Membaca File / Blob menjadi string base64 murni tanpa kompresi
 */
export async function fileToBase64(file: File | Blob): Promise<string> {
	return new Promise((resolve, reject) => {
		const reader = new FileReader();
		reader.onerror = () => reject(new Error('Gagal membaca file'));
		reader.onload = () => resolve(reader.result as string);
		reader.readAsDataURL(file);
	});
}

/**
 * Upload gambar ke Google Drive Bucket via Google Apps Script
 */
export async function uploadToGDrive(
	file: File | Blob,
	options: GDriveUploadOptions = {}
): Promise<GDriveUploadResult> {
	const shouldCompress = options.compress !== false && (file.type.startsWith('image/') || !file.type);
	let base64String = '';
	let mimeType = file.type || 'image/jpeg';

	try {
		if (shouldCompress) {
			const compressed = await compressImage(
				file,
				options.maxWidth ?? 1200,
				options.maxHeight ?? 1200,
				options.quality ?? 0.8
			);
			base64String = compressed.base64;
			mimeType = compressed.mimeType;
		} else {
			base64String = await fileToBase64(file);
		}

		let filename = options.filename;
		if (!filename) {
			if ('name' in file && typeof (file as File).name === 'string') {
				filename = (file as File).name;
			} else {
				filename = `upload_${Date.now()}.jpg`;
			}
		}

		// Pastikan ekstensi sesuai jika dikompres ke JPEG
		if (shouldCompress && !filename.match(/\.(jpe?g)$/i)) {
			filename = filename.replace(/\.[^/.]+$/, '') + '.jpg';
		}

		const payload = {
			action: 'upload',
			filename,
			mimeType,
			base64: base64String,
			folderName: options.folder || 'uploads'
		};

		// PENTING: Gunakan 'text/plain;charset=utf-8' agar browser tidak mengirim request OPTIONS (CORS preflight)
		// karena Google Apps Script tidak mendukung OPTIONS preflight.
		const response = await fetch(GDRIVE_UPLOAD_URL, {
			method: 'POST',
			headers: {
				'Content-Type': 'text/plain;charset=utf-8'
			},
			body: JSON.stringify(payload)
		});

		if (!response.ok) {
			throw new Error(`HTTP error! status: ${response.status}`);
		}

		const data = await response.json();
		return data as GDriveUploadResult;
	} catch (error: any) {
		console.error('Error saat mengunggah ke Google Drive Bucket:', error);
		return {
			status: 'error',
			message: error?.message || 'Gagal mengunggah gambar ke Google Drive.'
		};
	}
}

/**
 * Menghapus file dari Google Drive Bucket (dipindahkan ke Trash)
 * Dapat menerima File ID atau URL gambar langsung
 */
export async function deleteFromGDrive(
	fileIdOrUrl?: string | null
): Promise<{ status: 'success' | 'error'; message: string; fileId?: string }> {
	if (!fileIdOrUrl) {
		return { status: 'error', message: 'File ID atau URL kosong.' };
	}

	const fileId = extractGDriveFileId(fileIdOrUrl);
	// Jika bukan file Google Drive (misalnya gambar Unsplash / CDN luar), lewati tanpa error
	if (!fileId && !isGDriveUrl(fileIdOrUrl)) {
		return {
			status: 'success',
			message: 'Bukan file Google Drive, dilewati.'
		};
	}

	if (!fileId) {
		return {
			status: 'error',
			message: 'Format URL Google Drive tidak valid untuk ekstraksi ID.'
		};
	}

	console.log(`[GDrive Bucket] Memulai penghapusan file ID: ${fileId}`);

	// 1. Coba request via POST
	try {
		const payload = {
			action: 'delete',
			fileId: fileId
		};

		const response = await fetch(GDRIVE_UPLOAD_URL, {
			method: 'POST',
			headers: {
				'Content-Type': 'text/plain;charset=utf-8'
			},
			body: JSON.stringify(payload)
		});

		if (response.ok) {
			const data = await response.json();
			if (data && data.status === 'success') {
				console.log(`[GDrive Bucket] Berhasil menghapus file (POST):`, data);
				return data;
			}
		}
	} catch (postErr) {
		console.warn('[GDrive Bucket] Request POST delete gagal atau dialihkan, mencoba fallback GET:', postErr);
	}

	// 2. Fallback via GET parameter (lebih ramah redirect CORS di lingkungan browser tertentu)
	try {
		const getUrl = `${GDRIVE_UPLOAD_URL}?action=delete&fileId=${encodeURIComponent(fileId)}`;
		const getResponse = await fetch(getUrl, { method: 'GET' });
		if (getResponse.ok) {
			const getData = await getResponse.json();
			console.log(`[GDrive Bucket] Berhasil menghapus file (GET fallback):`, getData);
			return getData;
		}
	} catch (getErr: any) {
		console.error('[GDrive Bucket] Fallback GET delete juga mengalami error:', getErr);
	}

	return {
		status: 'error',
		message: 'Gagal menghubungi server Google Drive Apps Script untuk menghapus file.'
	};
}

/**
 * Healthcheck ping ke Apps Script Web App
 */
export async function pingGDriveBucket(): Promise<boolean> {
	try {
		const res = await fetch(GDRIVE_UPLOAD_URL, { method: 'GET' });
		if (!res.ok) return false;
		const data = await res.json();
		return data.status === 'success';
	} catch {
		return false;
	}
}
