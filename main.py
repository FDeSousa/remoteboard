from kivy.app import App
from kivy.lang import Builder
from kivy.uix.boxlayout import BoxLayout
from kivy.uix.button import Button
from kivy.uix.gridlayout import GridLayout
from kivy.properties import StringProperty, NumericProperty, ListProperty
from kivy.uix.screenmanager import ScreenManager, Screen
from kivy.properties import StringProperty


builder_string = """
# Defining the screen manager, even with only one screen
ScreenManager:
    id: screen_manager

    ButtonScreen:
        id: button_screen
        name: 'ButtonScreen'
        manager: screen_manager

# Widget definitions
# Rounded button
<RoundedButton@Button>:
    # Background colour must be invisible for this to work
    background_color: 0, 0, 0, 0
    border_radius: [25]

    canvas.before:
        Color:
            # Colours slightly different between states to show press
            rgba: (.4, .4, .4, 1) if self.state == 'normal' else (0, .7, .7, 1)
        RoundedRectangle:
            pos: self.pos
            size: self.size
            radius: self.border_radius

# Rounded image button
<RoundedImageButton@BoxLayout>:
    RoundedButton:
        text: root.text

        # Image:
            # source: 'button_image_0.jpg'
            # center_x: self.parent.center_x
            # center_y: self.parent.center_y


# Defining the screen, holding all the buttons. All of them.
<ButtonScreen>:
    orientation: 'vertical'

    BoxLayout:
        id: button_screen_layout
        orientation: 'vertical'

        # =================================================
        # Number of columns in the grid view
        BoxLayout:
            orientation: 'horizontal'
            size_hint: (1.0, None)
            height: sp(50)

            Label:
                text: 'Col Count'
                width: sp(150)

            Slider:
                id: grid_view_cols_slider
                min: 2
                max: 10
                step: 1
                value: 3
                size_hint: (None, 1.0)
                width: sp(200)
                height: sp(50)

            Label:
                text: str(grid_view_cols_slider.value)
                halign: 'right'
                size_hint: (None, 1.0)
                height: sp(50)
                text_size: self.size

        # =================================================
        # Number of rows in the grid view
        BoxLayout:
            orientation: 'horizontal'
            size_hint: (1.0, None)
            height: sp(50)

            Label:
                text: 'Row Count'
                width: sp(150)

            Slider:
                id: grid_view_rows_slider
                min: 2
                max: 10
                step: 1
                value: 3
                size_hint: (None, 1.0)
                width: sp(200)
                height: sp(50)

            Label:
                text: str(grid_view_rows_slider.value)
                halign: 'right'
                size_hint: (None, 1.0)
                height: sp(50)
                text_size: self.size

        # =================================================
        # Spacing and padding
        BoxLayout:
            orientation: 'horizontal'
            size_hint: (1.0, None)
            #width: sp(400)
            height: sp(50)

            Label:
                text: 'Spacing'

            Slider:
                id: grid_view_spacing_slider
                min: 0
                max: 25
                step: 1
                value: 5
                size_hint: (None, 1.0)
                width: sp(200)
                height: sp(50)

            Label:
                text: str(grid_view_spacing_slider.value)
                halign: 'right'
                size_hint: (None, 1.0)
                height: sp(50)
                text_size: self.size

        # =================================================
        # Grid of buttons
        GridLayout:
            id: grid_view_buttons
            cols: 1
            rows: 1
            #cols: grid_view_cols_slider.value
            #rows: grid_view_rows_slider.value
            spacing: (grid_view_spacing_slider.value, grid_view_spacing_slider.value)
            padding: (grid_view_spacing_slider.value, grid_view_spacing_slider.value)
"""


class ButtonScreen(Screen):
    pass


class RoundedButton(Button):
    pass


class RoundedImageButton(BoxLayout):
    text = StringProperty(None)
    #image = StringProperty(None)
    def __init__(self, text):
        super().__init__(text=text)


class RemoteBoard(App):
    def build(self):
        sm = ScreenManager()
        sm.add_widget(ButtonScreen(name=ButtonScreen.__name__))
        return sm

    def on_start(self, **kwargs):
        col_count_slider = self.root.current_screen.ids.grid_view_cols_slider
        col_count_slider.bind(value=self.on_cols_change)
        col_count_slider.value = 5

        row_count_slider = self.root.current_screen.ids.grid_view_rows_slider
        row_count_slider.bind(value=self.on_rows_change)
        row_count_slider.value = 5

    def on_cols_change(self, instance, value):
        grid = self.root.current_screen.ids.grid_view_buttons
        self.resize_grid(grid, value, grid.rows)

    def on_rows_change(self, instance, value):
        grid = self.root.current_screen.ids.grid_view_buttons
        self.resize_grid(grid, grid.cols, value)

    def resize_grid(self, grid, cols, rows):
        while len(grid.children):
            grid.remove_widget(grid.children[0])
        grid.cols = cols
        grid.rows = rows
        for n in range(cols * rows):
            grid.add_widget(RoundedImageButton(f'Button {n}'))


Builder.load_string(builder_string)


if __name__ == '__main__':
    RemoteBoard().run()
