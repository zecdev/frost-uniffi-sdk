import XCTest
@testable import FrostSwift

public class FrostSwiftTests: XCTest {
    func testFrost() {
        XCTAssertEqual(FrostSwift.frost(), "❄️")
    }
}
