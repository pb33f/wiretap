export function ExchangeMethod(method: string): string {
    switch (method) {
        case 'GET':
            return 'success'
        case 'POST':
            return 'primary'
        case 'PUT':
            return 'primary'
        case 'DELETE':
            return 'danger'
        case 'PATCH':
            return 'warning'
        default:
            return 'neutral'
    }
}