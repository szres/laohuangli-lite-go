<script>
	export let title;
	export let positive;
	export let negative;
</script>

<div class="inline-flex mt-1">
	<div class="text-right mr-1 self-center font-bold text-xl">{title}</div>
	<div class="self-center text-balance">
		<div class="text-md">
			<span class="text-sm mr-0.5 rounded-md p-0.5 bg-secondary text-black">宜</span>{positive}
		</div>
		<div class="text-md">
			<span class="text-sm mr-0.5 rounded-md p-0.5 bg-accent text-black">忌</span>{negative}
		</div>
	</div>
</div>
