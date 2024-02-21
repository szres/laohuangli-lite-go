
export const load = async () => {
    const fetchJSON = async () => {
        const res = await fetch(`http://laohuangli-bot/templates.json`)
        const data = await res.json()
        return data
    }

    return fetchJSON()
}
