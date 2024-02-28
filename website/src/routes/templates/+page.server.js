export const load = async () => {
	const fetchJSON = async () => {
		const res1 = await fetch(`http://` + import.meta.env.VITE_DATA_URL + `/templates.json`);
		const res2 = await fetch(`http://` + import.meta.env.VITE_DATA_URL + `/laohuangli.json`);
		const res3 = await fetch(`http://` + import.meta.env.VITE_DATA_URL + `/laohuangli-user.json`);
		const templates = await res1.json();
		const entrys = await res2.json();
		const entrysUser = await res3.json();
		return { templates, entrys, entrysUser };
	};

	return fetchJSON();
};
