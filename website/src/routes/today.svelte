<script>
	import Daily from './daily.svelte';
	function formatDate(date) {
		var d = new Date(date),
			month = '' + (d.getMonth() + 1),
			day = '' + d.getDate(),
			year = d.getFullYear();

		if (month.length < 2) month = '0' + month;
		if (day.length < 2) day = '0' + day;

		return [year, month, day].join('-');
	}
</script>

<div class="select-none text-3xl w-full text-center font-bold">今日</div>
<div class="w-full text-sm text-center">{formatDate(Date.now())}</div>
<div class="flex justify-center">
	<div class="flex flex-col mt-2 inline-flex">
		<span class="text-sm text-center">开发中...</span>
		<Daily title="穿搭" positive="穿衣 穿衣" negative="裸体 裸体" />
		<Daily title="饮食" positive="吃饭 吃饭" negative="绝食 绝食" />
		<Daily title="出行" positive="走路 走路" negative="爬行 爬行" />
	</div>
</div>
