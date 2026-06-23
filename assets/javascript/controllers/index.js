import { application } from "controllers/application"
import SpeedReaderController from "controllers/speed_reader_controller"
import ScrollableTablePatternsController from "controllers/scrollable_table_patterns_controller"
import WcagContrastController from "controllers/wcag_contrast_controller"

application.register("speed-reader", SpeedReaderController)
application.register("scrollable-table-patterns", ScrollableTablePatternsController)
application.register("wcag-contrast", WcagContrastController)
