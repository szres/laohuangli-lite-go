<script>
	export let title;
	export let positive;
	export let negative;
</script>

<div class="inline-flex mt-1">
	<div class="text-right mr-2 self-center min-w-max font-bold text-lg lg:text-2xl">{title}</div>
	<div class="self-center text-pretty">
		<div class="flex text-xs lg:text-base">
			<div class="inline-flex self-center text-sm mr-0.5 rounded-md p-0.5 bg-secondary text-black">
				宜
			</div>
			<div class="inline-flex">{positive}</div>
		</div>
		<div class="flex text-xs lg:text-base">
			<div class="inline-flex self-center text-sm mr-0.5 rounded-md p-0.5 bg-accent text-black">
				忌
			</div>
			<div class="inline-flex">{negative}</div>
		</div>
	</div>
</div>
