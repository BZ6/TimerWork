import asyncio
from playwright.async_api import async_playwright
import random
import string
import time

class TimerWorkE2ETests:
	def __init__(self):
		self.base_url = "http://localhost:3000"
		self.test_username = f"test_user_{self.generate_random_string()}"
		self.test_password = "TestPassword123!"

	def generate_random_string(self, length=8):
		return ''.join(random.choices(string.ascii_lowercase + string.digits, k=length))

	async def take_screenshot(self, page, name):
		await page.screenshot(path=f"{name}.png")
		print(f"Скриншот: {name}.png")

	async def debug_page_content(self, page):
		buttons = await page.locator("button").all()
		for i, button in enumerate(buttons):
			text = await button.text_content()

		inputs = await page.locator("input").all()
		for i, inp in enumerate(inputs):
			placeholder = await inp.get_attribute("placeholder")

	async def test_user_registration(self, page):
		try:
			await page.goto(self.base_url, wait_until='networkidle')
			await page.wait_for_timeout(3000)

			await self.take_screenshot(page, "01_main_page")
			await self.debug_page_content(page)

			register_btn = page.locator("button:has-text('Нет аккаунта? Зарегистрироваться')").first
			if await register_btn.is_visible(timeout=2000):
				await register_btn.click()
				await page.wait_for_timeout(2000)
				await self.take_screenshot(page, "02_after_register_click")
			else:
				return False

			username_input = page.locator("input[type='text']").first
			if await username_input.is_visible(timeout=2000):
				await username_input.fill(self.test_username)

			else:

				return False

			password_input = page.locator("input[type='password']").first
			if await password_input.is_visible(timeout=2000):
				await password_input.fill(self.test_password)
			else:
				return False

			submit_btn = page.locator("button[type='submit']").first
			if await submit_btn.is_visible(timeout=2000):
				await submit_btn.click()
				await page.wait_for_timeout(3000)
				await self.take_screenshot(page, "03_after_registration")
			else:
				return False

			success_indicators = [
				"text=успешна",
				"text=successful",
				"text=создан",
				"text=created"
			]

			registration_success = False
			for selector in success_indicators:
				element = page.locator(selector).first
				if await element.is_visible(timeout=2000):
					text = await element.text_content()
					registration_success = True
					break

			if registration_success:
				return True
			else:
				error_indicators = [
					".error",
					"[class*='error']",
					"text=ошибка",
					"text=error",
					"text=уже существует"
				]

				for selector in error_indicators:
					error_element = page.locator(selector).first
					if await error_element.is_visible(timeout=1000):
						error_text = await error_element.text_content()
						return False
				return True

		except Exception as e:
			await self.take_screenshot(page, "error_registration")
			return False

	async def test_user_login(self, page):
		try:
			if "localhost:3000" not in page.url:
				await page.goto(self.base_url, wait_until='networkidle')
				await page.wait_for_timeout(2000)

			await self.take_screenshot(page, "04_login_page")

			username_input = page.locator("input[type='text']").first
			if await username_input.is_visible(timeout=2000):
				await username_input.fill(self.test_username)
			else:
				return False

			password_input = page.locator("input[type='password']").first
			if await password_input.is_visible(timeout=2000):
				await password_input.fill(self.test_password)
			else:
				return False

			login_btn = page.locator("button:has-text('Войти')").first
			if await login_btn.is_visible(timeout=2000):
				await login_btn.click()
				await page.wait_for_timeout(3000)
				await self.take_screenshot(page, "05_after_login")
			else:
				return False

			success_indicators = [
				"button:has-text('Выйти')",
				"button:has-text('Начало недели')",
				"text=Таймер",
				"text=История недель"
			]

			login_success = False
			for selector in success_indicators:
				element = page.locator(selector).first
				if await element.is_visible(timeout=3000):
					text = await element.text_content()
					login_success = True
					break

			if login_success:
				return True
			else:
				error_indicators = [
					".error",
					"[class*='error']",
					"text=неверн",
					"text=invalid"
				]

				for selector in error_indicators:
					error_element = page.locator(selector).first
					if await error_element.is_visible(timeout=1000):
						error_text = await error_element.text_content()
						return False

				await self.debug_page_content(page)
				return False

		except Exception as e:
			await self.take_screenshot(page, "error_login")
			return False

	async def test_timer_functionality(self, page):
		try:
			logout_btn = page.locator("button:has-text('Выйти')").first
			if not await logout_btn.is_visible(timeout=2000):
				return False

			await self.take_screenshot(page, "06_before_timer")

			start_btn = page.locator("button:has-text('Начало недели')").first
			if await start_btn.is_visible(timeout=2000):
				await start_btn.click()
				await page.wait_for_timeout(2000)
				await self.take_screenshot(page, "07_after_start_click")
			else:
				return False

			goal_input = page.locator("input[type='number']").first
			if await goal_input.is_visible(timeout=2000):
				await goal_input.fill("60")

				confirm_btn = page.locator("button:has-text('Начать неделю')").first
				if await confirm_btn.is_visible(timeout=2000):
					await confirm_btn.click()
					await page.wait_for_timeout(2000)
				else:
					return False
			else:
				return False

			await self.take_screenshot(page, "08_timer_started")

			timer_found = False
			for tag in ["h1", "h2", "div", "span"]:
				elements = await page.locator(tag).all()
				for element in elements:
					text = await element.text_content()
					if text and ":" in text:
						# Проверяем что это время (содержит цифры и двоеточия)
						if any(c.isdigit() for c in text) and text.count(":") >= 2:
							timer_found = True
							break
				if timer_found:
					break

			history_btn = page.locator("button:has-text('История недель')").first
			if await history_btn.is_visible(timeout=2000):
				await history_btn.click()
				await page.wait_for_timeout(2000)
				await self.take_screenshot(page, "09_history_page")
			else:
				return False

			timer_btn = page.locator("button:has-text('Таймер')").first
			if await timer_btn.is_visible(timeout=2000):
				await timer_btn.click()
				await page.wait_for_timeout(2000)
				await self.take_screenshot(page, "10_back_to_timer")
			else:
				return False
			return True
		except Exception as e:
			await self.take_screenshot(page, "error_timer")
			return False

	async def run_all_tests(self):
		async with async_playwright() as p:
			browser = await p.chromium.launch(headless=False)
			page = await browser.new_page()

			try:
				test1_passed = await self.test_user_registration(page)
				test2_passed = await self.test_user_login(page)
				test3_passed = await self.test_timer_functionality(page)

				print(f"Тест 1 (Регистрация): {'ПРОЙДЕН' if test1_passed else 'ПРОВАЛЕН'}")
				print(f"Тест 2 (Вход в систему): {'ПРОЙДЕН' if test2_passed else 'ПРОВАЛЕН'}")
				print(f"Тест 3 (Работа таймера): {'ПРОЙДЕН' if test3_passed else 'ПРОВАЛЕН'}")

				if test1_passed and test2_passed and test3_passed:
					print("\nВСЕ ТЕСТЫ УСПЕШНО ПРОЙДЕНЫ!")
				else:
					print("\nПроверьте скриншоты для отладки")

				return test1_passed and test2_passed and test3_passed
			finally:
				await browser.close()

if __name__ == "__main__":
	tester = TimerWorkE2ETests()
	success = asyncio.run(tester.run_all_tests())
	exit(0 if success else 1)