import QRCode from 'qrcode';

/**
 * Generate QR Code as Data URL directly in browser for table self-order link
 */
export async function getTableQRCodeDataURL(qrToken: string): Promise<string> {
	const currentOrigin = typeof window !== 'undefined' ? window.location.origin : 'http://localhost:5173';
	const url = `${currentOrigin}/order?token=${qrToken}`;
	return await QRCode.toDataURL(url, {
		width: 300,
		margin: 2,
		color: {
			dark: '#1e293b',
			light: '#ffffff'
		}
	});
}
